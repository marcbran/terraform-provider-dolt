package provider

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &TableResource{}
var _ resource.ResourceWithImportState = &TableResource{}

func NewTableResource() resource.Resource {
	return &TableResource{}
}

type TableResource struct {
	db *sql.DB
}

type TableResourceModel struct {
	Database types.String `tfsdk:"database"`
	Name     types.String `tfsdk:"name"`
	Query    types.String `tfsdk:"query"`
	Columns  types.List   `tfsdk:"columns"`
}

func (m TableResourceModel) useQuery() string {
	return fmt.Sprintf("USE %s", m.Database.ValueString())
}

func (m TableResourceModel) createQuery() string {
	return m.Query.ValueString()
}

func (m TableResourceModel) readQuery() string {
	return fmt.Sprintf(`
		SELECT COLUMN_NAME, COLUMN_TYPE, COLUMN_KEY
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE `+"`"+`TABLE_SCHEMA`+"`"+` = '%s' AND `+"`"+`TABLE_NAME`+"`"+` = '%s'
		ORDER BY ORDINAL_POSITION`, m.Database.ValueString(), m.Name.ValueString())
}

func (m TableResourceModel) deleteQuery() string {
	return fmt.Sprintf("DROP TABLE %s", m.Name.ValueString())
}

func (r *TableResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_table"
}

func (r *TableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Table resource",

		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{
				MarkdownDescription: "Name of the database that contains the table",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the table, not confirming equality with table created by query",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "Query to create the table",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"columns": schema.ListNestedAttribute{
				MarkdownDescription: "Table columns",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"key": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (r *TableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create table, got error: %s", err))
		return
	}

	_, err = tx.ExecContext(ctx, data.useQuery())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create table, got error: %s", err))
		return
	}

	_, err = tx.ExecContext(ctx, data.createQuery())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create table, got error: %s", err))
		return
	}

	err = tx.Commit()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create table, got error: %s", err))
		return
	}

	if r.fillData(ctx, &data, resp.Diagnostics) {
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

	if r.fillData(ctx, &data, resp.Diagnostics) {
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

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete table, got error: %s", err))
		return
	}

	_, err = tx.ExecContext(ctx, data.useQuery())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete table, got error: %s", err))
		return
	}

	_, err = tx.ExecContext(ctx, data.deleteQuery())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete table, got error: %s", err))
		return
	}

	err = tx.Commit()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete table, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted a table")
}

func (r *TableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TableResource) fillData(ctx context.Context, data *TableResourceModel, diag diag.Diagnostics) bool {
	result, err := r.db.QueryContext(ctx, data.readQuery())
	if err != nil {
		diag.AddError("Client Error", fmt.Sprintf("Unable to read table, got error: %s", err))
		return true
	}
	if !result.Next() {
		diag.AddError("Client Error", fmt.Sprintf("Cannot find table with name %s", data.Name.ValueString()))
		return true
	}
	var columns []types.Object
	for result.Next() {
		var name, typ, key string
		err := result.Scan(&name, &typ, &key)
		if err != nil {
			diag.AddError("Client Error", fmt.Sprintf("Unable to read table, got error: %s", err))
			return true
		}
		col, diagnostics := types.ObjectValue(columnType.AttrTypes, map[string]attr.Value{
			"name": types.StringValue(name),
			"type": types.StringValue(typ),
			"key":  types.StringValue(key),
		})
		diag.Append(diagnostics...)
		columns = append(columns, col)
	}
	columnsList, diagnostics := types.ListValueFrom(ctx, columnType, columns)
	diag.Append(diagnostics...)
	data.Columns = columnsList
	return false
}
