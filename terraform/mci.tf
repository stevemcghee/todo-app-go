# Enable Fleet (GKE Hub) registration for both clusters

# Register Primary Cluster
resource "google_gke_hub_membership" "primary" {
  membership_id = "primary-cluster-membership"
  endpoint {
    gke_cluster {
      resource_link = "//container.googleapis.com/${google_container_cluster.primary.id}"
    }
  }
  depends_on = [
    google_project_service.gkehub_api,
    google_container_cluster.primary
  ]
}

# Register Secondary Cluster
resource "google_gke_hub_membership" "secondary" {
  membership_id = "secondary-cluster-membership"
  endpoint {
    gke_cluster {
      resource_link = "//container.googleapis.com/${google_container_cluster.secondary.id}"
    }
  }
  depends_on = [
    google_project_service.gkehub_api,
    google_container_cluster.secondary
  ]
}

# Enable Multi-Cluster Ingress Feature on the Fleet
resource "google_gke_hub_feature" "mci" {
  name = "multiclusteringress"
  location = "global"
  spec {
    multiclusteringress {
      config_membership = google_gke_hub_membership.primary.id
    }
  }
  depends_on = [
    google_project_service.multiclusteringress_api
  ]
}

resource "google_compute_global_address" "todo_app_global_ip" {
  name = "todo-app-global-ip"
}
