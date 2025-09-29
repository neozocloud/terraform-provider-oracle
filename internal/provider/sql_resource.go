// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-oracle/internal/oracle"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var _ resource.Resource = &SqlResource{}

func NewSqlResource() resource.Resource {
	return &SqlResource{}
}

// SqlResource defines the resource implementation.
type SqlResource struct {
	client *oracle.Client
}

// SqlResourceModel describes the resource data model.
type SqlResourceModel struct {
	Sql types.String `tfsdk:"sql"`
	ID  types.String `tfsdk:"id"`
}

func (r *SqlResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql"
}

func (r *SqlResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Oracle SQL resource",

		Attributes: map[string]schema.Attribute{
			"sql": schema.StringAttribute{
				MarkdownDescription: "SQL to execute",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "SQL identifier",
			},
		},
	}
}

func (r *SqlResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SqlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SqlResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.ExecuteSQL(data.Sql.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to execute sql, got error: %s", err))
		return
	}

	data.ID = types.StringValue(data.Sql.ValueString())

	tflog.Trace(ctx, "executed sql")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SqlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SqlResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// This resource is for executing arbitrary SQL statements. It is a "write-only" resource in that it only has a Create function.
	// The Read, Update, and Delete functions do nothing. This is by design.
	// It is intended for one-off SQL executions, not for managing state.
	// For managing resources with state, use the other resources in this provider.
	// It is recommended to use this resource with the "create_before_destroy" lifecycle meta-argument to ensure that
	// a new SQL statement is executed before the old one is "deleted" (which is a no-op).
}

func (r *SqlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This resource is not designed to be updated.
	// If you need to execute a different SQL statement, you should create a new resource.
}

func (r *SqlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// This resource is "write-only".
	// The "delete" function does nothing.
}

//resource "oracle_sql" "my_table" {
//  # The content of the SQL statement is the unique identifier for this resource.
//   # If you change this content, Terraform will "replace" the resource.
//   sql = "CREATE TABLE my_app_table (id NUMBER, name VARCHAR2(100))"
//
//   lifecycle {
//     # This is the magic that makes it work.
//     # When the 'sql' content changes, Terraform will create the new resource
//     # (running the new SQL) before "destroying" the old one (which is a no-op).
//     create_before_destroy = true
//   }
//}
