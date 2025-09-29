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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DirectoryResource{}
var _ resource.ResourceWithImportState = &DirectoryResource{}

func NewDirectoryResource() resource.Resource {
	return &DirectoryResource{}
}

// DirectoryResource defines the resource implementation.
type DirectoryResource struct {
	client *oracle.Client
}

// DirectoryResourceModel describes the resource data model.
type DirectoryResourceModel struct {
	Name types.String `tfsdk:"name"`
	Path types.String `tfsdk:"path"`
	ID   types.String `tfsdk:"id"`
}

func (r *DirectoryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_directory"
}

func (r *DirectoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Oracle Directory resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Directory name",
				Required:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Directory path",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Directory identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DirectoryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DirectoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DirectoryResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	directory := oracle.Directory{
		Name: data.Name.ValueString(),
		Path: data.Path.ValueString(),
	}

	err := r.client.CreateDirectory(directory)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create directory, got error: %s", err))
		return
	}

	data.ID = data.Name

	tflog.Trace(ctx, "created a directory resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DirectoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DirectoryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	directory, err := r.client.ReadDirectory(data.ID.ValueString())
	if err != nil {
		// If the directory is not found, remove it from state
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(strings.ToLower(directory.Name))
	data.Path = types.StringValue(directory.Path)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DirectoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DirectoryResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	directory := oracle.Directory{
		Name: data.Name.ValueString(),
		Path: data.Path.ValueString(),
	}

	err := r.client.CreateDirectory(directory)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update directory, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DirectoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DirectoryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DropDirectory(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to drop directory, got error: %s", err))
		return
	}
}

func (r *DirectoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
