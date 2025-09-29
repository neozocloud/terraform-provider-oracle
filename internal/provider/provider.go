// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neozocloud/terraform-provider-oracle/internal/oracle"
)

// Ensure OracleRDBMSProvider satisfies various provider interfaces.
var _ provider.Provider = &OracleRDBMSProvider{}
var _ provider.ProviderWithFunctions = &OracleRDBMSProvider{}
var _ provider.ProviderWithEphemeralResources = &OracleRDBMSProvider{}

// OracleRDBMSProvider defines the provider implementation.
type OracleRDBMSProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// OracleRDBMSProviderModel describes the provider data model.
type OracleRDBMSProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Port     types.String `tfsdk:"port"`
	Service  types.String `tfsdk:"service"`
}

func (p *OracleRDBMSProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "oracle"
	resp.Version = p.version
}

func (p *OracleRDBMSProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
			},
			"port": schema.StringAttribute{
				Optional: true,
			},
			"username": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"service": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (p *OracleRDBMSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config OracleRDBMSProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Oracle Host",
			"The provider cannot create the Oracle client as there is an unknown configuration value for the Oracle host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ORACLE_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Oracle Username",
			"The provider cannot create the Oracle client as there is an unknown configuration value for the Oracle username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ORACLE_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Oracle Password",
			"The provider cannot create the Oracle client as there is an unknown configuration value for the Oracle password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ORACLE_PASSWORD environment variable.",
		)
	}

	if config.Port.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Unknown Oracle Port",
			"The provider cannot create the Oracle client as there is an unknown configuration value for the Oracle port. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ORACLE_PORT environment variable.",
		)
	}

	if config.Service.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("service"),
			"Unknown Oracle Service",
			"The provider cannot create the Oracle client as there is an unknown configuration value for the Oracle service. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ORACLE_SERVICE environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("ORACLE_HOST")
	port := os.Getenv("ORACLE_PORT")
	username := os.Getenv("ORACLE_USERNAME")
	password := os.Getenv("ORACLE_PASSWORD")
	service := os.Getenv("ORACLE_SERVICE")
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Port.IsNull() {
		port = config.Port.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if !config.Service.IsNull() {
		service = config.Service.ValueString()
	}
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Oracle Host",
			"The provider cannot create the Oracle client as there is a missing or empty value for the Oracle host. "+
				"Set the host value in the configuration or use the ORACLE_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Oracle Username",
			"The provider cannot create the Oracle client as there is a missing or empty value for the Oracle username. "+
				"Set the username value in the configuration or use the ORACLE_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Oracle Password",
			"The provider cannot create the Oracle client as there is a missing or empty value for the Oracle password. "+
				"Set the password value in the configuration or use the ORACLE_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if port == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Missing Oracle Port",
			"The provider cannot create the Oracle client as there is a missing or empty value for the Oracle port. "+
				"Set the port value in the configuration or use the ORACLE_PORT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if service == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("service"),
			"Missing Oracle Service",
			"The provider cannot create the Oracle client as there is a missing or empty value for the Oracle service. "+
				"Set the service value in the configuration or use the ORACLE_SERVICE environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	dbPort, err := strconv.Atoi(port)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Port Value",
			"The provided port value is not a valid integer.",
		)
		return
	}

	client, err := oracle.NewClient(host, service, username, password, dbPort)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Oracle Client",
			"An unexpected error occurred when creating the Oracle client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Oracle Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OracleRDBMSProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewRoleResource,
		NewGrantSystemPrivilegesResource,
		NewGrantObjectPrivilegesResource,
		NewGrantDirectoryPrivilegesResource,
		NewDirectoryResource,
		NewSqlResource,
	}
}

func (p *OracleRDBMSProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *OracleRDBMSProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *OracleRDBMSProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OracleRDBMSProvider{
			version: version,
		}
	}
}
