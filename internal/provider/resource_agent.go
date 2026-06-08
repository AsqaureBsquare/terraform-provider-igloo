package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AgentResource{}

type AgentResource struct{}

func NewAgentResource() resource.Resource {
	return &AgentResource{}
}

type AgentResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Token             types.String `tfsdk:"token"`
	ServerURL         types.String `tfsdk:"server_url"`
	WorkspaceID       types.String `tfsdk:"workspace_id"`
	NetworkMonitoring types.Bool   `tfsdk:"network_monitoring"`
}

func (r *AgentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (r *AgentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Exposes the Igloo agent token and workspace metadata injected by the provisioner.

Use this resource to connect your workspace container to the Igloo control plane:

` + "```hcl\n" + `resource "igloo_agent" "main" {}

# Pass the agent token to your container:
resource "kubernetes_deployment" "workspace" {
  ...
  env {
    name  = "IGLOO_AGENT_TOKEN"
    value = igloo_agent.main.token
  }
  env {
    name  = "IGLOO_SERVER_URL"
    value = igloo_agent.main.server_url
  }
  env {
    name  = "IGLOO_WORKSPACE_ID"
    value = igloo_agent.main.workspace_id
  }
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The workspace ID (same as workspace_id, used as Terraform resource ID).",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "JWT agent token used by the workspace agent to authenticate with the Igloo server.",
			},
			"server_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "URL of the Igloo control plane server.",
			},
			"workspace_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Igloo workspace ID for this workspace.",
			},
			"network_monitoring": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable network connection monitoring for this workspace. When true, the agent collects TCP connection snapshots and exposes them via the Network Map UI. Defaults to false.",
			},
		},
	}
}

func (r *AgentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AgentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := os.Getenv("IGLOO_AGENT_TOKEN")
	serverURL := os.Getenv("IGLOO_SERVER_URL")
	workspaceID := os.Getenv("IGLOO_WORKSPACE_ID")

	if token == "" {
		resp.Diagnostics.AddError(
			"Missing IGLOO_AGENT_TOKEN",
			"The IGLOO_AGENT_TOKEN environment variable must be set. It is injected automatically by the Igloo provisioner.",
		)
		return
	}
	if serverURL == "" {
		resp.Diagnostics.AddError(
			"Missing IGLOO_SERVER_URL",
			"The IGLOO_SERVER_URL environment variable must be set. It is injected automatically by the Igloo provisioner.",
		)
		return
	}
	if workspaceID == "" {
		resp.Diagnostics.AddError(
			"Missing IGLOO_WORKSPACE_ID",
			"The IGLOO_WORKSPACE_ID environment variable must be set. It is injected automatically by the Igloo provisioner.",
		)
		return
	}

	data.ID = types.StringValue(workspaceID)
	data.Token = types.StringValue(token)
	data.ServerURL = types.StringValue(serverURL)
	data.WorkspaceID = types.StringValue(workspaceID)
	if data.NetworkMonitoring.IsNull() || data.NetworkMonitoring.IsUnknown() {
		data.NetworkMonitoring = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AgentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh from env vars each time — values are stable within a provisioning run.
	if token := os.Getenv("IGLOO_AGENT_TOKEN"); token != "" {
		data.Token = types.StringValue(token)
	}
	if serverURL := os.Getenv("IGLOO_SERVER_URL"); serverURL != "" {
		data.ServerURL = types.StringValue(serverURL)
	}
	if workspaceID := os.Getenv("IGLOO_WORKSPACE_ID"); workspaceID != "" {
		data.WorkspaceID = types.StringValue(workspaceID)
		data.ID = types.StringValue(workspaceID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AgentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgentResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Nothing to destroy — the agent resource is purely computed from env vars.
}
