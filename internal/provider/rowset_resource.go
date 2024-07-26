package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

var _ resource.Resource = &RowSetResource{}
var _ resource.ResourceWithImportState = &RowSetResource{}

func NewRowSetResource() resource.Resource {
	return &RowSetResource{}
}

// RowSetResource defines the resource implementation.
type RowSetResource struct {
	client *http.Client
}

// RowSetResourceModel describes the resource data model.
type RowSetResourceModel struct {
	RepositoryPath types.String `tfsdk:"repository_path"`
	TableName      types.String `tfsdk:"table_name"`
	UniqueColumn   types.String `tfsdk:"unique_column"`
	Columns        types.List   `tfsdk:"columns"`
	Values         types.Map    `tfsdk:"values"`
}

func (r *RowSetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rowset"
}

func (r *RowSetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "RowSet resource",

		Attributes: map[string]schema.Attribute{
			"repository_path": schema.StringAttribute{
				MarkdownDescription: "Path to the data repository that holds the row set",
				Required:            true,
			},
			"table_name": schema.StringAttribute{
				MarkdownDescription: "Name of the table where the set of rows will be stored",
				Required:            true,
			},
			"unique_column": schema.StringAttribute{
				MarkdownDescription: "Column that will be used to uniquely identify each row",
				Required:            true,
			},
			"columns": schema.ListAttribute{
				MarkdownDescription: "Columns for which values will be inserted",
				ElementType:         types.StringType,
				Required:            true,
			},
			"values": schema.MapAttribute{
				MarkdownDescription: "Values to be inserted into the table",
				ElementType:         types.ListType{ElemType: types.StringType},
				Required:            true,
			},
		},
	}
}

func (r *RowSetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RowSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RowSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repositoryPath := data.RepositoryPath.ValueString()
	abs, err := filepath.Abs(repositoryPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create row set, got error: %s", err))
		return
	}

	var columns []string
	for _, c := range data.Columns.Elements() {
		column := c.(basetypes.StringValue)
		columns = append(columns, column.ValueString())
	}
	columnsString := strings.Join(columns, ", ")
	var multipleValues []string
	for _, vs := range data.Values.Elements() {
		valuesList := vs.(basetypes.ListValue)
		var values []string
		for _, v := range valuesList.Elements() {
			values = append(values, v.String())
		}
		valuesString := fmt.Sprintf("(%s)", strings.Join(values, ", "))
		multipleValues = append(multipleValues, valuesString)
	}
	multipleValuesString := strings.Join(multipleValues, ", ")
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s;`, data.TableName.ValueString(), columnsString, multipleValuesString)

	cmd := exec.Command("dolt", "sql", "-q", query)
	cmd.Dir = abs
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create row set, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "created a resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RowSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RowSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RowSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RowSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RowSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RowSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repositoryPath := data.RepositoryPath.ValueString()
	abs, err := filepath.Abs(repositoryPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete row set, got error: %s", err))
		return
	}

	var uniqueValues []string
	for key := range data.Values.Elements() {
		uniqueValues = append(uniqueValues, fmt.Sprintf("\"%s\"", key))
	}
	uniqueValuesString := strings.Join(uniqueValues, ", ")
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s IN (%s);`, data.TableName.ValueString(), data.UniqueColumn.ValueString(), uniqueValuesString)

	cmd := exec.Command("dolt", "sql", "-q", query)
	cmd.Dir = abs
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete row set, got error: %s", err))
		return
	}
}

func (r *RowSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
