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

provider "kubernetes" {
  # Uses in-cluster config when running inside the provisioner pod.
  # The provisioner's service account must have permission to create
  # pods, pvcs, and deployments in the "workspaces" namespace.
}

# Expose the workspace agent token and metadata injected by Igloo.
resource "igloo_agent" "main" {}

variable "cpu" {
  default = "500m"
}

variable "memory" {
  default = "512Mi"
}

variable "disk" {
  default = "10Gi"
}

locals {
  labels = {
    "igloo.dev/workspace" = igloo_agent.main.workspace_id
  }
}

resource "kubernetes_persistent_volume_claim" "workspace" {
  metadata {
    name      = "ws-${igloo_agent.main.workspace_id}"
    namespace = "workspaces"
    labels    = local.labels
  }
  spec {
    access_modes       = ["ReadWriteOnce"]
    storage_class_name = "gp3"
    resources {
      requests = {
        storage = var.disk
      }
    }
  }
  wait_until_bound = false
}

resource "kubernetes_deployment" "workspace" {
  metadata {
    name      = "ws-${igloo_agent.main.workspace_id}"
    namespace = "workspaces"
    labels    = local.labels
  }
  spec {
    replicas = 1
    selector {
      match_labels = local.labels
    }
    template {
      metadata {
        labels = local.labels
      }
      spec {
        container {
          name              = "workspace"
          image             = "ghcr.io/asqaurebsquare/igloo-workspace:main"
          image_pull_policy = "Always"

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

          resources {
            requests = {
              cpu    = var.cpu
              memory = var.memory
            }
            limits = {
              cpu    = var.cpu
              memory = var.memory
            }
          }

          volume_mount {
            name       = "home"
            mount_path = "/home/user"
          }

          security_context {
            run_as_user  = 1000
            run_as_group = 1000
          }
        }

        volume {
          name = "home"
          persistent_volume_claim {
            claim_name = kubernetes_persistent_volume_claim.workspace.metadata[0].name
          }
        }
      }
    }
  }
}
