// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

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
var _ resource.Resource = &GrantObjectPrivilegesResource{}
var _ resource.ResourceWithImportState = &GrantObjectPrivilegesResource{}

func NewGrantObjectPrivilegesResource() resource.Resource {
	return &GrantObjectPrivilegesResource{}
}

// GrantObjectPrivilegesResource defines the resource implementation.
type GrantObjectPrivilegesResource struct {
	client *oracle.Client
}

// GrantObjectPrivilegesResourceModel describes the resource data model.
type GrantObjectPrivilegesResourceModel struct {
	Principal  types.String `tfsdk:"principal"`
	Object     types.String `tfsdk:"object"`
	Owner      types.String `tfsdk:"owner"`
	Privileges types.Set    `tfsdk:"privileges"`
	GrantsMode types.String `tfsdk:"grants_mode"`
	ID         types.String `tfsdk:"id"`
}

func (r *GrantObjectPrivilegesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_grant_object_privileges"
}

func (r *GrantObjectPrivilegesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Oracle Grant Object Privileges resource",

		Attributes: map[string]schema.Attribute{
			"principal": schema.StringAttribute{
				MarkdownDescription: "Principal",
				Required:            true,
			},
			"object": schema.StringAttribute{
				MarkdownDescription: "Object",
				Required:            true,
			},
			"owner": schema.StringAttribute{
				MarkdownDescription: "Owner",
				Optional:            true,
			},
			"privileges": schema.SetAttribute{
				MarkdownDescription: "Privileges",
				ElementType:         types.StringType,
				Required:            true,
			},
			"grants_mode": schema.StringAttribute{
				MarkdownDescription: "Grants mode",
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

func (r *GrantObjectPrivilegesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GrantObjectPrivilegesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GrantObjectPrivilegesResourceModel

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

	grant := oracle.ObjectPrivilege{
		Principal:  data.Principal.ValueString(),
		Object:     data.Object.ValueString(),
		Owner:      data.Owner.ValueString(),
		Privileges: privileges,
		GrantsMode: data.GrantsMode.ValueString(),
	}

	err := r.client.GrantObjectPrivileges(grant)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to grant object privileges, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", data.Principal.ValueString(), data.Owner.ValueString(), data.Object.ValueString()))

	tflog.Trace(ctx, "granted object privileges")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrantObjectPrivilegesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GrantObjectPrivilegesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(data.ID.ValueString(), ":")
	principal := parts[0]
	owner := parts[1]
	object := parts[2]

	privileges, err := r.client.GetCurrentObjectPrivileges(principal, owner, object)
	if err != nil {
		// If the grant is not found, remove it from state
		resp.State.RemoveResource(ctx)
		return
	}

	data.Principal = types.StringValue(principal)
	data.Owner = types.StringValue(owner)
	data.Object = types.StringValue(object)
	data.Privileges, resp.Diagnostics = types.SetValueFrom(ctx, types.StringType, privileges)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrantObjectPrivilegesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GrantObjectPrivilegesResourceModel

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

	grant := oracle.ObjectPrivilege{
		Principal:  data.Principal.ValueString(),
		Object:     data.Object.ValueString(),
		Owner:      data.Owner.ValueString(),
		Privileges: privileges,
		GrantsMode: data.GrantsMode.ValueString(),
	}

	err := r.client.GrantObjectPrivileges(grant)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update object privileges, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrantObjectPrivilegesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GrantObjectPrivilegesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	grant := oracle.ObjectPrivilege{
		Principal:  data.Principal.ValueString(),
		Object:     data.Object.ValueString(),
		Owner:      data.Owner.ValueString(),
		Privileges: []string{},
		GrantsMode: "enforce",
	}

	err := r.client.GrantObjectPrivileges(grant)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to revoke object privileges, got error: %s", err))
		return
	}
}

func (r *GrantObjectPrivilegesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
