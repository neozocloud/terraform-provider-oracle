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
var _ resource.Resource = &GrantRolesResource{}
var _ resource.ResourceWithImportState = &GrantRolesResource{}

func NewGrantRolesResource() resource.Resource {
	return &GrantRolesResource{}
}

// GrantRolesResource defines the resource implementation.
type GrantRolesResource struct {
	client *oracle.Client
}

// GrantRolesResourceModel describes the resource data model.
type GrantRolesResourceModel struct {
	Principal  types.String `tfsdk:"principal"`
	Roles      types.Set    `tfsdk:"roles"`
	GrantsMode types.String `tfsdk:"grants_mode"`
	ID         types.String `tfsdk:"id"`
}

func (r *GrantRolesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_grant_roles"
}

func (r *GrantRolesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A resource to manage roles for a user or role.",

		Attributes: map[string]schema.Attribute{
			"principal": schema.StringAttribute{
				MarkdownDescription: "The user or role to whom the roles are granted.",
				Required:            true,
			},
			"roles": schema.SetAttribute{
				MarkdownDescription: "The roles to grant to the principal. (This should be specified in lowercase. for example: `connect`, `resource`)",
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

func (r *GrantRolesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GrantRolesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GrantRolesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var roles []string
	resp.Diagnostics.Append(data.Roles.ElementsAs(ctx, &roles, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grant := oracle.GrantRole{
		Principal:  data.Principal.ValueString(),
		Roles:      roles,
		GrantsMode: data.GrantsMode.ValueString(),
	}

	err := r.client.GrantRoles(grant)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to grant roles, got error: %s", err))
		return
	}

	data.ID = data.Principal

	tflog.Trace(ctx, "granted roles")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrantRolesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GrantRolesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	roles, err := r.client.GetCurrentRoles(data.ID.ValueString())
	if err != nil {
		// If the grant is not found, remove it from the state
		resp.State.RemoveResource(ctx)
		return
	}

	data.Principal = data.ID
	data.Roles, resp.Diagnostics = types.SetValueFrom(ctx, types.StringType, roles)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrantRolesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GrantRolesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var roles []string
	resp.Diagnostics.Append(data.Roles.ElementsAs(ctx, &roles, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grant := oracle.GrantRole{
		Principal:  data.Principal.ValueString(),
		Roles:      roles,
		GrantsMode: data.GrantsMode.ValueString(),
	}

	err := r.client.GrantRoles(grant)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update roles, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrantRolesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GrantRolesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var roles []string
	resp.Diagnostics.Append(data.Roles.ElementsAs(ctx, &roles, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grant := oracle.GrantRole{
		Principal: data.Principal.ValueString(),
		Roles:     roles,
	}

	err := r.client.RevokeRoles(grant)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to revoke roles, got error: %s", err))
		return
	}
}

func (r *GrantRolesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
