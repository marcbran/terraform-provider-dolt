package provider

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/dolthub/driver"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &DoltProvider{}
var _ provider.ProviderWithFunctions = &DoltProvider{}

type DoltProvider struct {
	version string
}

type DoltProviderModel struct {
	Path  types.String `tfsdk:"path"`
	Name  types.String `tfsdk:"name"`
	Email types.String `tfsdk:"email"`
}

func (m DoltProviderModel) databaseUrl() string {
	return fmt.Sprintf("file://%s?commitname=%s&commitemail=%s", m.Path.ValueString(), m.Name.ValueString(), m.Email.ValueString())
}

func (p *DoltProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dolt"
	resp.Version = p.version
}

func (p *DoltProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				MarkdownDescription: "Path to the directory where your databases are on disk",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the committer seen in the dolt commit log",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email of the committer seen in the dolt commit log",
				Required:            true,
			},
		},
	}
}

func (p *DoltProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data DoltProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	db, err := sql.Open("dolt", data.databaseUrl())
	if err != nil {
		return
	}

	_, err = db.Exec("SET @@dolt_transaction_commit=1")
	if err != nil {
		return
	}

	resp.DataSourceData = db
	resp.ResourceData = db
}

func (p *DoltProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDatabaseResource,
		NewTableResource,
		NewViewResource,
		NewRowSetResource,
	}
}

func (p *DoltProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *DoltProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DoltProvider{
			version: version,
		}
	}
}
