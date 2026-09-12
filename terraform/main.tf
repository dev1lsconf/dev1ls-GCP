# Enable required GCP APIs (No cost to enable APIs)
resource "google_project_service" "run_api" {
  project                    = var.project_id
  service                    = "run.googleapis.com"
  disable_on_destroy         = false
  disable_dependent_services = false
}

resource "google_project_service" "artifactregistry_api" {
  project                    = var.project_id
  service                    = "artifactregistry.googleapis.com"
  disable_on_destroy         = false
  disable_dependent_services = false
}

# Artifact Registry Repository for Docker images (Free tier: 0.5 GB/month)
resource "google_artifact_registry_repository" "repo" {
  project       = var.project_id
  location      = var.region
  repository_id = "devops-repo"
  description   = "Docker repository for DevOps prototype"
  format        = "DOCKER"

  depends_on = [google_project_service.artifactregistry_api]
}

# Principle of Least Privilege: Dedicated runtime service account with no admin rights
resource "google_service_account" "cloudrun_sa" {
  project      = var.project_id
  account_id   = "cloudrun-runtime-sa"
  display_name = "Cloud Run Runtime SA"
}

# Cloud Run Service (Always Free tier: 2M requests/month, 360k GB-s)
resource "google_cloud_run_v2_service" "api" {
  name     = var.service_name
  location = var.region
  project  = var.project_id
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.cloudrun_sa.email

    scaling {
      min_instance_count = 0 # Scale to 0 when idle = $0 cost!
      max_instance_count = 2 # Guardrail against traffic spikes
    }

    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.repo.repository_id}/${var.service_name}:${var.image_tag}"

      resources {
        limits = {
          cpu    = "1"
          memory = "128Mi"
        }
      }

      ports {
        container_port = 8080
      }

      env {
        name  = "APP_ENV"
        value = "production"
      }

      env {
        name  = "APP_VERSION"
        value = var.image_tag
      }
    }
  }

  depends_on = [google_project_service.run_api]
}

# Allow unauthenticated access (Public Web API)
resource "google_cloud_run_v2_service_iam_member" "public_access" {
  project  = var.project_id
  location = google_cloud_run_v2_service.api.location
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
