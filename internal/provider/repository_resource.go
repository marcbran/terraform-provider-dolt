package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &RepositoryResource{}
var _ resource.ResourceWithImportState = &RepositoryResource{}

func NewRepositoryResource() resource.Resource {
	return &RepositoryResource{}
}

// RepositoryResource defines the resource implementation.
type RepositoryResource struct {
	client *http.Client
}

// RepositoryResourceModel describes the resource data model.
type RepositoryResourceModel struct {
	Path  types.String `tfsdk:"path"`
	Name  types.String `tfsdk:"name"`
	Email types.String `tfsdk:"email"`
}

func (r *RepositoryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (r *RepositoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Repository resource",

		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				MarkdownDescription: "Path to the repository",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Author name",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Author email",
				Required:            true,
			},
		},
	}
}

func (r *RepositoryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *RepositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RepositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := data.Path.ValueString()
	abs, err := filepath.Abs(p)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create repository, got error: %s", err))
		return
	}
	err = os.MkdirAll(p, 0755)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create repository, got error: %s", err))
		return
	}

	cmd := exec.Command("dolt", "init", "--name", data.Name.ValueString(), "--email", data.Email.ValueString())
	cmd.Dir = abs
	err = cmd.Run()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create repository, got error: %s", err))
		return
	}

	// For the purposes of this repository code, hardcoding a response value to
	// save into the Terraform state.
	//data.Id = types.StringValue("repository-id")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RepositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RepositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RepositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RepositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RepositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RepositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := data.Path.ValueString()
	abs, err := filepath.Abs(p)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete repository, got error: %s", err))
		return
	}
	err = os.RemoveAll(filepath.Join(abs, ".dolt"))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete repository, got error: %s", err))
		return
	}
	err = os.Remove(abs)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete repository, got error: %s", err))
		return
	}
}

func (r *RepositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func commitQuery(message string, authorName string, authorEmail string) string {
	return fmt.Sprintf("CALL DOLT_COMMIT('-A', '-m', '%s', '--author', '%s <%s>');", message, authorName, authorEmail)
}

func execQuery(dir string, query string) error {
	var stderr strings.Builder
	cmd := exec.Command("dolt", "sql", "-q", query)
	cmd.Dir = dir
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Unable to execute query, got error: %s\n\n%s", err, stderr.String())
	}
	return nil
}
