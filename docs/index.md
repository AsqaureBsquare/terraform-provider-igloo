---
page_title: "Igloo Provider"
description: |-
  Use the Igloo provider to interact with Igloo workspace resources from inside a provisioner.
---

# Igloo Provider

The Igloo provider is used inside Igloo workspace provisioners — Terraform templates that run inside the Igloo platform to create and manage developer workspaces.

When Igloo launches a provisioner, it injects three environment variables:

| Variable | Description |
|---|---|
| `IGLOO_AGENT_TOKEN` | JWT token the workspace agent uses to authenticate back to the server |
| `IGLOO_SERVER_URL` | Base URL of the Igloo server (e.g. `https://igloos.cloud`) |
| `IGLOO_WORKSPACE_ID` | Unique ID of the workspace being provisioned |

The `igloo_agent` resource reads these variables and exposes them as Terraform outputs so you can pass them into the workspace container.

## Example Usage

```terraform
terraform {
  required_providers {
    igloo = {
      source  = "AsqaureBsquare/igloo"
      version = "~> 0.1"
    }
  }
}

provider "igloo" {}

resource "igloo_agent" "main" {}
```

## Provider Configuration

The Igloo provider requires no configuration — all values are read from environment variables injected by the Igloo platform at provisioner runtime.
