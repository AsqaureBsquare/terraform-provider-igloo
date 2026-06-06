---
page_title: "igloo_agent Resource - Igloo"
description: |-
  Reads the workspace agent token and metadata injected by the Igloo platform.
---

# igloo_agent (Resource)

Reads the workspace agent token and metadata injected by the Igloo platform into the provisioner environment. Use this resource to pass agent credentials into the workspace container.

This resource is read-only — it only surfaces values that Igloo has already provided. Destroying it is a no-op.

## Example Usage

### Kubernetes workspace

```terraform
terraform {
  required_providers {
    igloo = {
      source  = "AsqaureBsquare/igloo"
      version = "~> 0.1"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

provider "igloo" {}
provider "kubernetes" {}

resource "igloo_agent" "main" {}

resource "kubernetes_deployment" "workspace" {
  metadata {
    name      = "ws-${igloo_agent.main.workspace_id}"
    namespace = "workspaces"
  }
  spec {
    replicas = 1
    selector {
      match_labels = { "igloo.dev/workspace" = igloo_agent.main.workspace_id }
    }
    template {
      metadata {
        labels = { "igloo.dev/workspace" = igloo_agent.main.workspace_id }
      }
      spec {
        container {
          name  = "workspace"
          image = "ghcr.io/asqaurebsquare/igloo-workspace:main"

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
      }
    }
  }
}
```

## Schema

### Read-Only

- `id` (String) — Same as `workspace_id`. Used as the Terraform resource ID.
- `token` (String, Sensitive) — Agent JWT token. Pass this as `IGLOO_AGENT_TOKEN` in the workspace container.
- `server_url` (String) — Base URL of the Igloo server. Pass this as `IGLOO_SERVER_URL` in the workspace container.
- `workspace_id` (String) — Unique ID of the workspace being provisioned. Pass this as `IGLOO_WORKSPACE_ID` in the workspace container.
