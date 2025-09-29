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

	"terraform-provider-oracle/internal/oracle"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

// UserResource defines the resource implementation.
type UserResource struct {
	client *oracle.Client
}

// UserResourceModel describes the resource data model.
type UserResourceModel struct {
	Username              types.String `tfsdk:"username"`
	Password              types.String `tfsdk:"password"`
	DefaultTablespace     types.String `tfsdk:"default_tablespace"`
	DefaultTempTablespace types.String `tfsdk:"default_temp_tablespace"`
	Profile               types.String `tfsdk:"profile"`
	AuthenticationType    types.String `tfsdk:"authentication_type"`
	State                 types.String `tfsdk:"state"`
	ID                    types.String `tfsdk:"id"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Oracle User resource",

		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "Username",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password",
				Optional:            true,
				Sensitive:           true,
			},
			"default_tablespace": schema.StringAttribute{
				MarkdownDescription: "Default tablespace",
				Optional:            true,
				Computed:            true,
			},
			"default_temp_tablespace": schema.StringAttribute{
				MarkdownDescription: "Default temporary tablespace",
				Optional:            true,
				Computed:            true,
			},
			"profile": schema.StringAttribute{
				MarkdownDescription: "Profile",
				Optional:            true,
				Computed:            true,
			},
			"authentication_type": schema.StringAttribute{
				MarkdownDescription: "Authentication type",
				Optional:            true,
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Account state",
				Optional:            true,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.AuthenticationType.ValueString() == "" {
		data.AuthenticationType = types.StringValue("password")
	}

	user := oracle.User{
		Username:              data.Username.ValueString(),
		Password:              data.Password.ValueString(),
		DefaultTablespace:     data.DefaultTablespace.ValueString(),
		DefaultTempTablespace: data.DefaultTempTablespace.ValueString(),
		Profile:               data.Profile.ValueString(),
		AuthenticationType:    data.AuthenticationType.ValueString(),
		State:                 data.State.ValueString(),
	}

	err := r.client.CreateUser(user)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user, got error: %s", err))
		return
	}

	data.ID = data.Username

	tflog.Trace(ctx, "created a user resource")

	// Read back user details to populate computed fields
	createdUser, err := r.client.ReadUser(user.Username)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}

	data.DefaultTablespace = types.StringValue(createdUser.DefaultTablespace)
	data.DefaultTempTablespace = types.StringValue(createdUser.DefaultTempTablespace)
	data.Profile = types.StringValue(createdUser.Profile)
	data.AuthenticationType = types.StringValue(createdUser.AuthenticationType)
	data.State = types.StringValue(createdUser.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.ReadUser(data.ID.ValueString())
	if err != nil {
		// If the user is not found, remove it from state
		resp.State.RemoveResource(ctx)
		return
	}

	data.Username = types.StringValue(user.Username)
	data.DefaultTablespace = types.StringValue(user.DefaultTablespace)
	data.DefaultTempTablespace = types.StringValue(user.DefaultTempTablespace)
	data.Profile = types.StringValue(user.Profile)
	data.AuthenticationType = types.StringValue(user.AuthenticationType)
	data.State = types.StringValue(user.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user := oracle.User{
		Username:              data.Username.ValueString(),
		Password:              data.Password.ValueString(),
		DefaultTablespace:     data.DefaultTablespace.ValueString(),
		DefaultTempTablespace: data.DefaultTempTablespace.ValueString(),
		Profile:               data.Profile.ValueString(),
		AuthenticationType:    data.AuthenticationType.ValueString(),
		State:                 data.State.ValueString(),
	}

	err := r.client.ModifyUser(user)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user, got error: %s", err))
		return
	}

	// Read back user details to populate computed fields
	updatedUser, err := r.client.ReadUser(user.Username)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}

	data.DefaultTablespace = types.StringValue(updatedUser.DefaultTablespace)
	data.DefaultTempTablespace = types.StringValue(updatedUser.DefaultTempTablespace)
	data.Profile = types.StringValue(updatedUser.Profile)
	data.AuthenticationType = types.StringValue(updatedUser.AuthenticationType)
	data.State = types.StringValue(updatedUser.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DropUser(data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to drop user, got error: %s", err))
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
