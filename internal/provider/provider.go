// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	Nne      types.String `tfsdk:"nne"`
}

const (
	nneEnabled  = "enabled"
	nneDisabled = "disabled"
	defaultPort = 1521
)

func resolveString(configured types.String, envVar string) string {
	if !configured.IsNull() {
		return configured.ValueString()
	}
	return os.Getenv(envVar)
}

func (p *OracleRDBMSProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "oracle"
	resp.Version = p.version
}

func (p *OracleRDBMSProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Oracle RDBMS provider to manage users, roles, and privileges.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "host name or IP address of the Oracle database server.",
				Optional:            true,
			},
			"port": schema.StringAttribute{
				Description: fmt.Sprintf("Listener port. Defaults to %d.", defaultPort),
				Optional:    true,
			},
			"nne": schema.StringAttribute{
				MarkdownDescription: "Oracle Native Network Encryption. Must be \"enabled\" or \"disabled\" " +
					"(default). When \"enabled\", the provider fails if the session is not encrypted.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(nneEnabled, nneDisabled),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "username to connect to the Oracle database server.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "password to connect to the Oracle database server.",
				Optional:            true,
				Sensitive:           true,
			},
			"service": schema.StringAttribute{
				MarkdownDescription: "service name of the Oracle database server.",
				Optional:            true,
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

	if config.Nne.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("nne"),
			"Unknown Oracle Nne",
			"The provider cannot create the Oracle client as there is an unknown configuration value for the Oracle NNE (NativeNetworkEncryption). "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ORACLE_NNE environment variable.",
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

	// Config values override environment variables.
	host := resolveString(config.Host, "ORACLE_HOST")
	username := resolveString(config.Username, "ORACLE_USERNAME")
	password := resolveString(config.Password, "ORACLE_PASSWORD")
	service := resolveString(config.Service, "ORACLE_SERVICE")
	portStr := resolveString(config.Port, "ORACLE_PORT")
	nne := resolveString(config.Nne, "ORACLE_NNE")

	// Required values must come from either the config or the environment.
	required := []struct {
		attr, envVar, value string
	}{
		{"host", "ORACLE_HOST", host},
		{"username", "ORACLE_USERNAME", username},
		{"password", "ORACLE_PASSWORD", password},
		{"service", "ORACLE_SERVICE", service},
	}
	for _, r := range required {
		if r.value == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root(r.attr),
				"Missing Oracle provider configuration",
				fmt.Sprintf("Set %q in the provider block or via the %s environment variable.", r.attr, r.envVar),
			)
		}
	}

	// Port: optional, defaults to 1521.
	port := defaultPort
	if portStr != "" {
		parsed, err := strconv.Atoi(portStr)
		if err != nil || parsed < 1 || parsed > 65535 {
			resp.Diagnostics.AddAttributeError(
				path.Root("port"),
				"Invalid port",
				fmt.Sprintf("%q is not a valid TCP port (1-65535).", portStr),
			)
		} else {
			port = parsed
		}
	}

	// NNE: optional, defaults to disabled. The schema validator only covers
	// the provider block, so values from the environment are checked here.
	switch nne {
	case "":
		nne = nneDisabled
	case nneEnabled, nneDisabled:
	default:
		resp.Diagnostics.AddAttributeError(
			path.Root("nne"),
			"Invalid nne value",
			fmt.Sprintf("%q is not valid; must be %q or %q.", nne, nneEnabled, nneDisabled),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := oracle.NewClient(host, service, username, password, port, nne == nneEnabled)
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
		NewGrantRolesResource,
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
