// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neozocloud/terraform-provider-oracle/internal/oracle"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var _ resource.Resource = &GrantSystemPrivilegesResource{}
var _ resource.ResourceWithImportState = &GrantSystemPrivilegesResource{}

func NewGrantSystemPrivilegesResource() resource.Resource {
	return &GrantSystemPrivilegesResource{}
}

// GrantSystemPrivilegesResource defines the resource implementation.
type GrantSystemPrivilegesResource struct {
	client *oracle.Client
}

// GrantSystemPrivilegesResourceModel describes the resource data model.
type GrantSystemPrivilegesResourceModel struct {
	Principal  types.String `tfsdk:"principal"`
	Privileges types.Set    `tfsdk:"privileges"`
	GrantsMode types.String `tfsdk:"grants_mode"`
	ID         types.String `tfsdk:"id"`
}

func (r *GrantSystemPrivilegesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_grant_system_privileges"
}

func (r *GrantSystemPrivilegesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A resource to manage system privileges for a user or role.",

		Attributes: map[string]schema.Attribute{
			"principal": schema.StringAttribute{
				MarkdownDescription: "The user or role to whom the privileges are granted.",
				Required:            true,
			},
			"privileges": schema.SetAttribute{
				MarkdownDescription: "The system privileges to grant to the principal. (This should be specified in uppercase. for example: `CREATE SESSION`)",
				ElementType:         types.StringType,
				Required:            true,
			},
			"grants_mode": schema.StringAttribute{
				MarkdownDescription: "The grants mode to use. If not specified, the default is `append`.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Grant identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *GrantSystemPrivilegesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*oracle.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *oracle.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *GrantSystemPrivilegesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GrantSystemPrivilegesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var privileges []string
	resp.Diagnostics.Append(data.Privileges.ElementsAs(ctx, &privileges, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grant := oracle.Grant{
		Principal:  data.Principal.ValueString(),
		Privileges: privileges,
		GrantsMode: data.GrantsMode.ValueString(),
	}

	err := r.client.GrantSystemPrivileges(grant)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to grant system privileges, got error: %s", err))
		return
	}

	data.ID = data.Principal

	tflog.Trace(ctx, "granted system privileges")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrantSystemPrivilegesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GrantSystemPrivilegesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	privileges, err := r.client.GetCurrentSystemPrivileges(data.ID.ValueString())
	if err != nil {
		// If the grant is not found, remove it from the state
		resp.State.RemoveResource(ctx)
		return
	}

	data.Principal = data.ID
	data.Privileges, resp.Diagnostics = types.SetValueFrom(ctx, types.StringType, privileges)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrantSystemPrivilegesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GrantSystemPrivilegesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var privileges []string
	resp.Diagnostics.Append(data.Privileges.ElementsAs(ctx, &privileges, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grant := oracle.Grant{
		Principal:  data.Principal.ValueString(),
		Privileges: privileges,
		GrantsMode: data.GrantsMode.ValueString(),
	}

	err := r.client.GrantSystemPrivileges(grant)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update system privileges, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrantSystemPrivilegesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GrantSystemPrivilegesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	grant := oracle.Grant{
		Principal:  data.Principal.ValueString(),
		Privileges: []string{},
		GrantsMode: "enforce",
	}

	err := r.client.GrantSystemPrivileges(grant)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to revoke system privileges, got error: %s", err))
		return
	}
}

func (r *GrantSystemPrivilegesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
