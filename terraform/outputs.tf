output "artifact_registry_repo" {
  description = "Artifact Registry Docker repository URL"
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.repo.repository_id}"
}

output "service_url" {
  description = "Public HTTPS URL for the deployed Cloud Run service"
  value       = google_cloud_run_v2_service.api.uri
}
