package provider

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &TableResource{}
var _ resource.ResourceWithImportState = &TableResource{}

func NewTableResource() resource.Resource {
	return &TableResource{}
}

// TableResource defines the resource implementation.
type TableResource struct {
	client *http.Client
}

// TableResourceModel describes the resource data model.
type TableResourceModel struct {
	RepositoryPath types.String `tfsdk:"repository_path"`
	Name           types.String `tfsdk:"name"`
	Query          types.String `tfsdk:"query"`
}

func (r *TableResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_table"
}

func (r *TableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Table resource",

		Attributes: map[string]schema.Attribute{
			"repository_path": schema.StringAttribute{
				MarkdownDescription: "Path to the data repository that holds the table",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the table, not confirming equality with table created by query",
				Required:            true,
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "Query to create the table",
				Required:            true,
			},
		},
	}
}

func (r *TableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repositoryPath := data.RepositoryPath.ValueString()
	abs, err := filepath.Abs(repositoryPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create table, got error: %s", err))
		return
	}
	cmd := exec.Command("dolt", "sql", "-q", data.Query.ValueString())
	cmd.Dir = abs
	err = cmd.Run()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create table, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "created a resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repositoryPath := data.RepositoryPath.ValueString()
	abs, err := filepath.Abs(repositoryPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create table, got error: %s", err))
		return
	}
	cmd := exec.Command("dolt", "sql", "-q", "DROP TABLE "+data.Name.ValueString())
	cmd.Dir = abs
	err = cmd.Run()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create table, got error: %s", err))
		return
	}
}

func (r *TableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
