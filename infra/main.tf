terraform {
  required_version = ">= 1.5"

  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = "0.18.0" # fijada: el repo del provider esta archivado
    }
    render = {
      source  = "render-oss/render"
      version = "1.9.1"
    }
  }
}

provider "neon" {}   # lee NEON_API_KEY del entorno
provider "render" {} # lee RENDER_API_KEY y RENDER_OWNER_ID del entorno

# --- El proyecto de Neon de PREPROD: no existe hasta que esto se aplica ---



resource "neon_role" "preprod" {
  project_id = neon_project.preprod.id
  branch_id  = neon_project.preprod.default_branch_id # default_branch_id, NO branch_id
  name       = "app_preprod_owner"
}

resource "neon_database" "preprod" {
  project_id = neon_project.preprod.id
  branch_id  = neon_project.preprod.default_branch_id
  name       = "app_preprod"
  owner_name = neon_role.preprod.name
}

resource "neon_project" "preprod" {
  name   = "mallki-preprod"
  org_id = "org-purple-river-77024456" # tu organizacion de Neon (no es secreto)

  history_retention_seconds = 21600
}

locals {
  # La cadena de conexion se ARMA aca con lo que devolvio Neon: nadie la copia y pega.
  # Formato de lib/pq (Go): postgres://usuario:clave@host/base?sslmode=require
  conn = "postgres://${neon_role.preprod.name}:${urlencode(neon_role.preprod.password)}@${neon_project.preprod.database_host}/${neon_database.preprod.name}?sslmode=require"
}

resource "render_web_service" "api" {
  name   = "mallki-api-preprod"
  plan   = "free"
  region = var.region

  runtime_source = {
    image = {
      image_url = var.image_repo_api # SIN la etiqueta pegada
      tag       = var.image_tag      # la etiqueta va en su propio campo
    }
  }

  env_vars = {
    DATABASE_URL = { value = local.conn }
  }
}

resource "render_web_service" "front" {
  name   = "mallki-front-preprod"
  plan   = "free"
  region = var.region

  runtime_source = {
    image = {
      image_url = var.image_repo_front
      tag       = var.image_tag
    }
  }

  env_vars = {
    BACKEND_URL  = { value = render_web_service.api.url } # la URL sale de Terraform, no de un copy-paste
    DNS_RESOLVER = { value = "8.8.8.8" }
  }
}