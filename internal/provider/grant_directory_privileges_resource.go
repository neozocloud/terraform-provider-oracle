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

	"terraform-provider-oracle/internal/oracle"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var _ resource.Resource = &GrantDirectoryPrivilegesResource{}
var _ resource.ResourceWithImportState = &GrantDirectoryPrivilegesResource{}

func NewGrantDirectoryPrivilegesResource() resource.Resource {
	return &GrantDirectoryPrivilegesResource{}
}

// GrantDirectoryPrivilegesResource defines the resource implementation.
type GrantDirectoryPrivilegesResource struct {
	client *oracle.Client
}

// GrantDirectoryPrivilegesResourceModel describes the resource data model.
type GrantDirectoryPrivilegesResourceModel struct {
	Principal  types.String `tfsdk:"principal"`
	Directory  types.String `tfsdk:"directory"`
	Privileges types.Set    `tfsdk:"privileges"`
	GrantsMode types.String `tfsdk:"grants_mode"`
	ID         types.String `tfsdk:"id"`
}

func (r *GrantDirectoryPrivilegesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_grant_directory_privileges"
}

func (r *GrantDirectoryPrivilegesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Oracle Grant Directory Privileges resource",

		Attributes: map[string]schema.Attribute{
			"principal": schema.StringAttribute{
				MarkdownDescription: "Principal",
				Required:            true,
			},
			"directory": schema.StringAttribute{
				MarkdownDescription: "Directory",
				Required:            true,
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

func (r *GrantDirectoryPrivilegesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GrantDirectoryPrivilegesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GrantDirectoryPrivilegesResourceModel

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

	grant := oracle.DirectoryPrivilege{
		Principal:  data.Principal.ValueString(),
		Directory:  data.Directory.ValueString(),
		Privileges: privileges,
		GrantsMode: data.GrantsMode.ValueString(),
	}

	err := r.client.GrantDirectoryPrivileges(grant)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to grant directory privileges, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", data.Principal.ValueString(), data.Directory.ValueString()))

	tflog.Trace(ctx, "granted directory privileges")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrantDirectoryPrivilegesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GrantDirectoryPrivilegesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(data.ID.ValueString(), ":")
	principal := parts[0]
	directory := parts[1]

	privileges, err := r.client.GetCurrentDirectoryPrivileges(principal, directory)
	if err != nil {
		// If the grant is not found, remove it from state
		resp.State.RemoveResource(ctx)
		return
	}

	data.Principal = types.StringValue(principal)
	data.Directory = types.StringValue(directory)
	data.Privileges, resp.Diagnostics = types.SetValueFrom(ctx, types.StringType, privileges)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrantDirectoryPrivilegesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GrantDirectoryPrivilegesResourceModel

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

	grant := oracle.DirectoryPrivilege{
		Principal:  data.Principal.ValueString(),
		Directory:  data.Directory.ValueString(),
		Privileges: privileges,
		GrantsMode: data.GrantsMode.ValueString(),
	}

	err := r.client.GrantDirectoryPrivileges(grant)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update directory privileges, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrantDirectoryPrivilegesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GrantDirectoryPrivilegesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	grant := oracle.DirectoryPrivilege{
		Principal:  data.Principal.ValueString(),
		Directory:  data.Directory.ValueString(),
		Privileges: []string{},
		GrantsMode: "enforce",
	}

	err := r.client.GrantDirectoryPrivileges(grant)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to revoke directory privileges, got error: %s", err))
		return
	}
}

func (r *GrantDirectoryPrivilegesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
