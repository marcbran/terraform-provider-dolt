package provider

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &TableDataSource{}

func NewTableDataSource() datasource.DataSource {
	return &TableDataSource{}
}

type TableDataSource struct {
	db *sql.DB
}

type TableDataSourceModel struct {
	Database types.String `tfsdk:"database"`
	Name     types.String `tfsdk:"name"`
	Columns  types.List   `tfsdk:"columns"`
}

var columnType = basetypes.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name": types.StringType,
		"type": types.StringType,
		"key":  types.StringType,
	},
}

func (m TableDataSourceModel) readQuery() string {
	return fmt.Sprintf(`
		SELECT COLUMN_NAME, COLUMN_TYPE, COLUMN_KEY
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE `+"`"+`TABLE_SCHEMA`+"`"+` = '%s' AND `+"`"+`TABLE_NAME`+"`"+` = '%s'
		ORDER BY ORDINAL_POSITION`, m.Database.ValueString(), m.Name.ValueString())
}

func (d *TableDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_table"
}

func (d *TableDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Table data source",

		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{
				MarkdownDescription: "Name of the database that contains the table",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the table",
				Required:            true,
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

func (d *TableDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.db = db
}

func (d *TableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TableDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.db.QueryContext(ctx, data.readQuery())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read table, got error: %s", err))
		return
	}
	if !result.Next() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Cannot find table with name %s", data.Name.ValueString()))
		return
	}
	var columns []types.Object
	for result.Next() {
		var name, typ, key string
		err := result.Scan(&name, &typ, &key)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read table, got error: %s", err))
			return
		}
		col, diagnostics := types.ObjectValue(columnType.AttrTypes, map[string]attr.Value{
			"name": types.StringValue(name),
			"type": types.StringValue(typ),
			"key":  types.StringValue(key),
		})
		resp.Diagnostics.Append(diagnostics...)
		columns = append(columns, col)
	}
	columnsList, diagnostics := types.ListValueFrom(ctx, columnType, columns)
	resp.Diagnostics.Append(diagnostics...)
	data.Columns = columnsList

	tflog.Trace(ctx, "read a data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
