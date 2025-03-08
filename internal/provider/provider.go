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
	"os"
	"path/filepath"
)

var _ provider.Provider = &DoltProvider{}
var _ provider.ProviderWithFunctions = &DoltProvider{}

type DoltProvider struct {
	version string
}

// DoltProviderModel TODO maybe path should be part of a resource rather than a provider config?
type DoltProviderModel struct {
	Path  types.String `tfsdk:"path"`
	Name  types.String `tfsdk:"name"`
	Email types.String `tfsdk:"email"`
}

func (m DoltProviderModel) databaseUrl() (string, error) {
	path, err := filepath.Abs(m.Path.ValueString())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file://%s?commitname=%s&commitemail=%s", path, m.Name.ValueString(), m.Email.ValueString()), nil
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

	_, err := os.Stat(data.Path.ValueString())
	if os.IsNotExist(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to configure provider, path doesn't exist: %s", data.Path.ValueString()))
		return
	}

	url, err := data.databaseUrl()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to configure provider, cannot produce database url for path: %s", data.Path.ValueString()))
		return
	}

	db, err := sql.Open("dolt", url)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to configure provider, cannot open database: %s", err))
		return
	}

	_, err = db.ExecContext(ctx, "SET @@dolt_transaction_commit=1")
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
	return []func() datasource.DataSource{
		NewDatabaseDataSource,
		NewTableDataSource,
	}
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
