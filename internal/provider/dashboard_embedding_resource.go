package provider

import (
	"context"
	"fmt"
	"strconv"

	"terraform-provider-superset/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &dashboardEmbeddingResource{}
	_ resource.ResourceWithConfigure   = &dashboardEmbeddingResource{}
	_ resource.ResourceWithImportState = &dashboardEmbeddingResource{}
)

func NewDashboardEmbeddingResource() resource.Resource {
	return &dashboardEmbeddingResource{}
}

type dashboardEmbeddingResource struct {
	client *client.Client
}

type dashboardEmbeddingResourceModel struct {
	ID             types.String `tfsdk:"id"`
	DashboardID    types.Int64  `tfsdk:"dashboard_id"`
	AllowedDomains types.List   `tfsdk:"allowed_domains"`
	UUID           types.String `tfsdk:"uuid"`
}

func (r *dashboardEmbeddingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard_embedding"
}

func (r *dashboardEmbeddingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages embedded dashboard configuration in Apache Superset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (dashboard_id).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dashboard_id": schema.Int64Attribute{
				Description: "The ID of the dashboard to embed.",
				Required:    true,
			},
			"allowed_domains": schema.ListAttribute{
				Description: "List of domains allowed to embed the dashboard. Use empty list to allow all domains.",
				Required:    true,
				ElementType: types.StringType,
			},
			"uuid": schema.StringAttribute{
				Description: "The UUID of the embedded dashboard. Use this to construct the embed URL.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *dashboardEmbeddingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData))
		return
	}
	r.client = c
}

func (r *dashboardEmbeddingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dashboardEmbeddingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var domains []string
	resp.Diagnostics.Append(plan.AllowedDomains.ElementsAs(ctx, &domains, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	embedded, err := r.client.CreateDashboardEmbedded(plan.DashboardID.ValueInt64(), domains)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create dashboard embedding", err.Error())
		return
	}

	plan.ID = types.StringValue(strconv.FormatInt(plan.DashboardID.ValueInt64(), 10))
	plan.UUID = types.StringValue(embedded.UUID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardEmbeddingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dashboardEmbeddingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	embedded, err := r.client.GetDashboardEmbedded(state.DashboardID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read dashboard embedding", err.Error())
		return
	}
	if embedded == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.UUID = types.StringValue(embedded.UUID)
	domains, diags := types.ListValueFrom(ctx, types.StringType, embedded.AllowedDomains)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.AllowedDomains = domains

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dashboardEmbeddingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dashboardEmbeddingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var domains []string
	resp.Diagnostics.Append(plan.AllowedDomains.ElementsAs(ctx, &domains, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	embedded, err := r.client.CreateDashboardEmbedded(plan.DashboardID.ValueInt64(), domains)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update dashboard embedding", err.Error())
		return
	}

	plan.ID = types.StringValue(strconv.FormatInt(plan.DashboardID.ValueInt64(), 10))
	plan.UUID = types.StringValue(embedded.UUID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardEmbeddingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dashboardEmbeddingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteDashboardEmbedded(state.DashboardID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Failed to delete dashboard embedding", err.Error())
	}
}

func (r *dashboardEmbeddingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	dashboardID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected numeric dashboard ID")
		return
	}

	embedded, err := r.client.GetDashboardEmbedded(dashboardID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import dashboard embedding", err.Error())
		return
	}
	if embedded == nil {
		resp.Diagnostics.AddError("Not found", fmt.Sprintf("No embedded config for dashboard %d", dashboardID))
		return
	}

	domains, diags := types.ListValueFrom(ctx, types.StringType, embedded.AllowedDomains)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := dashboardEmbeddingResourceModel{
		ID:             types.StringValue(strconv.FormatInt(dashboardID, 10)),
		DashboardID:    types.Int64Value(dashboardID),
		AllowedDomains: domains,
		UUID:           types.StringValue(embedded.UUID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
