package provider

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

type RowSetResource struct {
	db *sql.DB
}

type RowSetResourceModel struct {
	Database     types.String `tfsdk:"database"`
	Table        types.String `tfsdk:"table"`
	UniqueColumn types.String `tfsdk:"unique_column"`
	Columns      types.List   `tfsdk:"columns"`
	Values       types.Map    `tfsdk:"values"`
	RowCount     types.Int64  `tfsdk:"row_count"`
}

func (m RowSetResourceModel) useQuery() string {
	return fmt.Sprintf("USE %s", m.Database.ValueString())
}

func (m RowSetResourceModel) upsertQuery() string {
	var columns []string
	for _, c := range m.Columns.Elements() {
		if column, ok := c.(basetypes.StringValue); ok {
			columns = append(columns, column.ValueString())
		}
	}
	columnsString := strings.Join(columns, ", ")
	var multipleValues []string
	for _, vs := range m.Values.Elements() {
		if valuesList, ok := vs.(basetypes.ListValue); ok {
			var values []string
			for _, v := range valuesList.Elements() {
				values = append(values, v.String())
			}
			valuesString := fmt.Sprintf("(%s)", strings.Join(values, ", "))
			multipleValues = append(multipleValues, valuesString)
		}
	}
	multipleValuesString := strings.Join(multipleValues, ", ")
	var updateColumns []string
	for _, c := range m.Columns.Elements() {
		if column, ok := c.(basetypes.StringValue); ok {
			updateColumns = append(updateColumns, fmt.Sprintf("%s = VALUES(%s)", column.ValueString(), column.ValueString()))
		}
	}
	updateColumnsString := strings.Join(updateColumns, ", ")
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s ON DUPLICATE KEY UPDATE %s;`,
		m.Table.ValueString(), columnsString, multipleValuesString, updateColumnsString)
	return query
}

func (m RowSetResourceModel) pruneQuery(state RowSetResourceModel) string {
	var uniqueValues []string
	for key := range state.Values.Elements() {
		if _, ok := m.Values.Elements()[key]; !ok {
			uniqueValues = append(uniqueValues, fmt.Sprintf("\"%s\"", key))
		}
	}
	if len(uniqueValues) == 0 {
		return ""
	}
	uniqueValuesString := strings.Join(uniqueValues, ", ")
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s IN (%s);`,
		m.Table.ValueString(), m.UniqueColumn.ValueString(), uniqueValuesString)
	return query
}

func (m RowSetResourceModel) deleteQuery() string {
	var uniqueValues []string
	for key := range m.Values.Elements() {
		uniqueValues = append(uniqueValues, fmt.Sprintf("\"%s\"", key))
	}
	uniqueValuesString := strings.Join(uniqueValues, ", ")
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s IN (%s);`,
		m.Table.ValueString(), m.UniqueColumn.ValueString(), uniqueValuesString)
	return query
}

func (r *RowSetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rowset"
}

func (r *RowSetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "RowSet resource",

		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{
				MarkdownDescription: "Name of the database that contains the row set",
				Required:            true,
			},
			"table": schema.StringAttribute{
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
			"row_count": schema.Int64Attribute{
				MarkdownDescription: "Number of rows that are managed by this resource",
				Computed:            true,
			},
		},
	}
}

func (r *RowSetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	db, ok := req.ProviderData.(*sql.DB)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sql.DB, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.db = db
}

func (r *RowSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RowSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create row set, got error: %s", err))
		return
	}

	_, err = tx.Exec(data.useQuery())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create row set, got error: %s", err))
		return
	}

	_, err = tx.Exec(data.upsertQuery())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create row set, got error: %s", err))
		return
	}

	err = tx.Commit()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create row set, got error: %s", err))
		return
	}

	data.RowCount = types.Int64Value(int64(len(data.Values.Elements())))

	tflog.Trace(ctx, "created a row set")
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
	var data, state RowSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update row set, got error: %s", err))
		return
	}

	_, err = tx.Exec(data.useQuery())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update row set, got error: %s", err))
		return
	}

	_, err = tx.Exec(data.upsertQuery())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update row set, got error: %s", err))
		return
	}

	pruneQuery := data.pruneQuery(state)
	if pruneQuery != "" {
		_, err = tx.Exec(pruneQuery)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update row set, got error: %s", err))
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update row set, got error: %s", err))
		return
	}

	data.RowCount = types.Int64Value(int64(len(data.Values.Elements())))

	tflog.Trace(ctx, "updated a row set")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RowSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RowSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete row set, got error: %s", err))
		return
	}

	_, err = tx.Exec(data.useQuery())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete row set, got error: %s", err))
		return
	}

	_, err = tx.Exec(data.deleteQuery())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete row set, got error: %s", err))
		return
	}

	err = tx.Commit()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete row set, got error: %s", err))
		return
	}

	data.RowCount = types.Int64Value(0)

	tflog.Trace(ctx, "deleted a row set")
}

func (r *RowSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
