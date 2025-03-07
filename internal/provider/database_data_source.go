package provider

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &DatabaseDataSource{}

func NewDatabaseDataSource() datasource.DataSource {
	return &DatabaseDataSource{}
}

type DatabaseDataSource struct {
	db *sql.DB
}

type DatabaseDataSourceModel struct {
	Name types.String `tfsdk:"name"`
}

func (m DatabaseDataSourceModel) readQuery() string {
	return fmt.Sprintf("SHOW DATABASES LIKE '%s'", m.Name.ValueString())
}

func (d *DatabaseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (d *DatabaseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Database data source",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Database name",
				Required:            true,
			},
		},
	}
}

func (d *DatabaseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatabaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatabaseDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.db.QueryContext(ctx, data.readQuery())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read database, got error: %s", err))
		return
	}
	if !result.Next() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Cannot find database with name %s", data.Name.ValueString()))
		return
	}
	for result.Next() {
		var name string
		err := result.Scan(&name)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read database, got error: %s", err))
			return
		}
		if name != data.Name.ValueString() {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Expected database name %s to be %s", name, data.Name.ValueString()))
			return
		}
	}

	tflog.Trace(ctx, "read a data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
