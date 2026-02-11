package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-superset/internal/client"
)

var (
	_ resource.Resource                = &rowLevelSecurityResource{}
	_ resource.ResourceWithConfigure   = &rowLevelSecurityResource{}
	_ resource.ResourceWithImportState = &rowLevelSecurityResource{}
)

func NewRowLevelSecurityResource() resource.Resource {
	return &rowLevelSecurityResource{}
}

type rowLevelSecurityResource struct {
	client *client.Client
}

type rowLevelSecurityResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Tables      types.List   `tfsdk:"tables"`
	Clause      types.String `tfsdk:"clause"`
	RoleIDs     types.List   `tfsdk:"role_ids"`
	GroupKey    types.String `tfsdk:"group_key"`
	FilterType  types.String `tfsdk:"filter_type"`
	Description types.String `tfsdk:"description"`
}

func (r *rowLevelSecurityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_row_level_security"
}

func (r *rowLevelSecurityResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages row level security rules in Superset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the RLS rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the RLS rule.",
				Required:    true,
			},
			"tables": schema.ListAttribute{
				Description: "List of table/dataset IDs to apply RLS to.",
				ElementType: types.Int64Type,
				Required:    true,
			},
			"clause": schema.StringAttribute{
				Description: "SQL WHERE clause for row level security.",
				Required:    true,
			},
			"role_ids": schema.ListAttribute{
				Description: "List of role IDs to apply this RLS rule to.",
				ElementType: types.Int64Type,
				Optional:    true,
			},
			"group_key": schema.StringAttribute{
				Description: "Group key for RLS rule.",
				Optional:    true,
			},
			"filter_type": schema.StringAttribute{
				Description: "Filter type: 'Regular' or 'Base'. Defaults to 'Regular'.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the RLS rule.",
				Optional:    true,
			},
		},
	}
}

func (r *rowLevelSecurityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan rowLevelSecurityResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tables []int64
	diags = plan.Tables.ElementsAs(ctx, &tables, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating RLS rule", map[string]interface{}{
		"tables": tables,
	})

	var roleIDs []int64
	if !plan.RoleIDs.IsNull() {
		diags = plan.RoleIDs.ElementsAs(ctx, &roleIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	filterType := "Regular"
	if !plan.FilterType.IsNull() {
		filterType = plan.FilterType.ValueString()
	}

	rlsID, err := r.client.CreateRowLevelSecurity(
		plan.Name.ValueString(),
		tables,
		plan.Clause.ValueString(),
		roleIDs,
		plan.GroupKey.ValueString(),
		filterType,
		plan.Description.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating RLS rule",
			"Could not create RLS rule: "+err.Error(),
		)
		return
	}

	plan.ID = types.Int64Value(rlsID)
	if plan.FilterType.IsNull() {
		plan.FilterType = types.StringValue("Regular")
	}

	tflog.Debug(ctx, "Created RLS rule", map[string]interface{}{
		"id": rlsID,
	})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *rowLevelSecurityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state rowLevelSecurityResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rls, err := r.client.GetRowLevelSecurity(state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading RLS rule",
			"Could not read RLS rule ID "+fmt.Sprintf("%d", state.ID.ValueInt64())+": "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(rls.Name)
	state.Clause = types.StringValue(rls.Clause)
	state.GroupKey = types.StringValue(rls.GroupKey)
	state.FilterType = types.StringValue(rls.FilterType)
	state.Description = types.StringValue(rls.Description)

	tables, diags := types.ListValueFrom(ctx, types.Int64Type, rls.Tables)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Tables = tables

	roleIDs, diags := types.ListValueFrom(ctx, types.Int64Type, rls.RoleIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.RoleIDs = roleIDs

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *rowLevelSecurityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan rowLevelSecurityResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tables []int64
	diags = plan.Tables.ElementsAs(ctx, &tables, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var roleIDs []int64
	if !plan.RoleIDs.IsNull() {
		diags = plan.RoleIDs.ElementsAs(ctx, &roleIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	filterType := "Regular"
	if !plan.FilterType.IsNull() {
		filterType = plan.FilterType.ValueString()
	}

	err := r.client.UpdateRowLevelSecurity(
		plan.ID.ValueInt64(),
		plan.Name.ValueString(),
		tables,
		plan.Clause.ValueString(),
		roleIDs,
		plan.GroupKey.ValueString(),
		filterType,
		plan.Description.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating RLS rule",
			"Could not update RLS rule ID "+fmt.Sprintf("%d", plan.ID.ValueInt64())+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Updated RLS rule", map[string]interface{}{
		"id": plan.ID.ValueInt64(),
	})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *rowLevelSecurityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state rowLevelSecurityResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRowLevelSecurity(state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting RLS rule",
			"Could not delete RLS rule: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted RLS rule", map[string]interface{}{
		"id": state.ID.ValueInt64(),
	})
}

func (r *rowLevelSecurityResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *rowLevelSecurityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing RLS rule",
			"Could not parse RLS rule ID: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
