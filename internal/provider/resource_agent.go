package provider

import (
	"context"
	"fmt"
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
	InitScript        types.String `tfsdk:"init_script"`
}

func (r *AgentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (r *AgentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Exposes the Igloo agent token and workspace metadata injected by the provisioner.

Use ` + "`init_script`" + ` to start the agent in any container image — no agent pre-baked required:

` + "```hcl\n" + `resource "igloo_agent" "main" {}

resource "kubernetes_deployment" "workspace" {
  ...
  container {
    image   = var.workspace_image   # any Docker image
    command = ["sh", "-c", igloo_agent.main.init_script]
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
				MarkdownDescription: "Enable network connection monitoring. When true, the agent collects TCP connection snapshots and exposes them via the Network Map UI. Defaults to false.",
			},
			"init_script": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Shell script that downloads and starts the Igloo agent. Use as the container command to support any Docker image without pre-baking the agent binary.",
			},
		},
	}
}

func buildInitScript(serverURL, token, workspaceID string, networkMonitoring bool) string {
	netmap := "false"
	if networkMonitoring {
		netmap = "true"
	}
	return fmt.Sprintf(`#!/bin/sh
set -e
export IGLOO_SERVER_URL=%q
export IGLOO_AGENT_TOKEN=%q
export IGLOO_WORKSPACE_ID=%q
export IGLOO_NETMAP_ENABLED=%q
curl -fsSL "$IGLOO_SERVER_URL/downloads/igloo-agent-linux-amd64" -o /tmp/igloo-agent
chmod +x /tmp/igloo-agent
exec /tmp/igloo-agent`, serverURL, token, workspaceID, netmap)
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
		resp.Diagnostics.AddError("Missing IGLOO_AGENT_TOKEN",
			"The IGLOO_AGENT_TOKEN environment variable must be set. It is injected automatically by the Igloo provisioner.")
		return
	}
	if serverURL == "" {
		resp.Diagnostics.AddError("Missing IGLOO_SERVER_URL",
			"The IGLOO_SERVER_URL environment variable must be set. It is injected automatically by the Igloo provisioner.")
		return
	}
	if workspaceID == "" {
		resp.Diagnostics.AddError("Missing IGLOO_WORKSPACE_ID",
			"The IGLOO_WORKSPACE_ID environment variable must be set. It is injected automatically by the Igloo provisioner.")
		return
	}

	if data.NetworkMonitoring.IsNull() || data.NetworkMonitoring.IsUnknown() {
		data.NetworkMonitoring = types.BoolValue(false)
	}

	data.ID = types.StringValue(workspaceID)
	data.Token = types.StringValue(token)
	data.ServerURL = types.StringValue(serverURL)
	data.WorkspaceID = types.StringValue(workspaceID)
	data.InitScript = types.StringValue(buildInitScript(serverURL, token, workspaceID, data.NetworkMonitoring.ValueBool()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AgentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := os.Getenv("IGLOO_AGENT_TOKEN")
	serverURL := os.Getenv("IGLOO_SERVER_URL")
	workspaceID := os.Getenv("IGLOO_WORKSPACE_ID")

	if token != "" {
		data.Token = types.StringValue(token)
	}
	if serverURL != "" {
		data.ServerURL = types.StringValue(serverURL)
	}
	if workspaceID != "" {
		data.WorkspaceID = types.StringValue(workspaceID)
		data.ID = types.StringValue(workspaceID)
	}

	// Recompute init_script whenever any input changes.
	t := data.Token.ValueString()
	s := data.ServerURL.ValueString()
	w := data.WorkspaceID.ValueString()
	if t != "" && s != "" && w != "" {
		data.InitScript = types.StringValue(buildInitScript(s, t, w, data.NetworkMonitoring.ValueBool()))
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
