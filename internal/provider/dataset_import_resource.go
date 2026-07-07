package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-superset/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource               = &datasetImportResource{}
	_ resource.ResourceWithConfigure  = &datasetImportResource{}
	_ resource.ResourceWithModifyPlan = &datasetImportResource{}
)

func NewDatasetImportResource() resource.Resource {
	return &datasetImportResource{}
}

type datasetImportResource struct {
	client *client.Client
}

type datasetImportResourceModel struct {
	ID                types.String `tfsdk:"id"`
	SourceDir         types.String `tfsdk:"source_dir"`
	ForceOverwrite    types.Bool   `tfsdk:"force_overwrite"`
	DatabaseSecrets   types.Map    `tfsdk:"database_secrets"`
	DatabaseOverrides types.Map    `tfsdk:"database_overrides"`
	FileHashes        types.Map    `tfsdk:"file_hashes"`
}

func (r *datasetImportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset_import"
}

func (r *datasetImportResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Imports deduplicated datasets from one or more dashboard export directories via POST /api/v1/dataset/import/. " +
			"This endpoint properly respects overwrite=true, unlike the dashboard import endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this resource (derived from source_dir).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_dir": schema.StringAttribute{
				Description: "Path to a parent directory containing one or more dashboard export subdirectories. " +
					"Each subdirectory should contain datasets/, databases/, etc.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"force_overwrite": schema.BoolAttribute{
				Description: "Whether to overwrite existing datasets on import. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"database_secrets": schema.MapAttribute{
				Description: "Map of database UUID to database password/secret.",
				Optional:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
			"database_overrides": schema.MapAttribute{
				Description: "Map of database UUID to a JSON-encoded object of YAML field overrides.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"file_hashes": schema.MapAttribute{
				Description: "Map of deduplicated file path to SHA256 hash. Changes trigger re-import.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *datasetImportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *datasetImportResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan datasetImportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceDir := plan.SourceDir.ValueString()
	if sourceDir == "" {
		return
	}

	overrides := parseDatabaseOverrides(ctx, plan.DatabaseOverrides)
	collected, err := collectDedupedFiles(sourceDir, []string{"datasets", "databases"}, overrides)
	if err != nil {
		resp.Diagnostics.AddWarning("Cannot compute file hashes", err.Error())
		return
	}

	newHashes := hashCollectedFiles(collected)

	if req.State.Raw.IsNull() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("file_hashes"), toStringMap(newHashes))...)
		return
	}

	var state datasetImportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	oldHashes := fromStringMap(state.FileHashes)
	if !mapsEqual(oldHashes, newHashes) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("file_hashes"), toStringMap(newHashes))...)
	} else {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("file_hashes"), state.FileHashes)...)
	}
}

func (r *datasetImportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan datasetImportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.doImport(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Failed to import datasets", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *datasetImportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state datasetImportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Dataset import is a bulk operation — no remote object to verify.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datasetImportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan datasetImportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state datasetImportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	if err := r.doImport(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Failed to re-import datasets", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *datasetImportResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Datasets are not deleted when this resource is removed — they remain in Superset.
}

func (r *datasetImportResource) doImport(ctx context.Context, plan *datasetImportResourceModel) error {
	sourceDir := plan.SourceDir.ValueString()
	plan.ID = types.StringValue(fmt.Sprintf("dataset-import:%s", sourceDir))

	overrides := parseDatabaseOverrides(ctx, plan.DatabaseOverrides)

	collected, err := collectDedupedFiles(sourceDir, []string{"datasets", "databases"}, overrides)
	if err != nil {
		return fmt.Errorf("collecting dataset files: %w", err)
	}

	plan.FileHashes = toStringMap(hashCollectedFiles(collected))

	secrets := make(map[string]string)
	if !plan.DatabaseSecrets.IsNull() && !plan.DatabaseSecrets.IsUnknown() {
		diags := plan.DatabaseSecrets.ElementsAs(ctx, &secrets, false)
		if diags.HasError() {
			return fmt.Errorf("reading database_secrets")
		}
	}

	passwordMap := buildPasswordMapFromCollected(collected, secrets)
	passwords := ""
	if len(passwordMap) > 0 {
		b, _ := json.Marshal(passwordMap)
		passwords = string(b)
	}

	zipData, err := buildImportZip(collected, "SqlaTable", sourceDir)
	if err != nil {
		return fmt.Errorf("building ZIP: %w", err)
	}

	overwrite := plan.ForceOverwrite.ValueBool()
	tflog.Info(ctx, fmt.Sprintf("Importing %d deduplicated dataset files from %s (overwrite=%v)", len(collected), sourceDir, overwrite))

	if err := r.client.ImportDataset(zipData, overwrite, passwords); err != nil {
		return err
	}

	return nil
}
