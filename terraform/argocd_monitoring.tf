resource "kubernetes_manifest" "argocd_server_monitoring" {
  manifest = {
    apiVersion = "monitoring.googleapis.com/v1"
    kind       = "PodMonitoring"
    metadata = {
      name      = "argocd-monitoring"
      namespace = "argocd"
    }
    spec = {
      selector = {
        matchLabels = {
          "app.kubernetes.io/name" = "argocd-server"
        }
      }
      endpoints = [
        {
          port     = 8083
          path     = "/metrics"
          interval = "30s"
        }
      ]
    }
  }
  depends_on = [helm_release.argocd]
}

resource "kubernetes_manifest" "argocd_repo_server_monitoring" {
  manifest = {
    apiVersion = "monitoring.googleapis.com/v1"
    kind       = "PodMonitoring"
    metadata = {
      name      = "argocd-repo-server-monitoring"
      namespace = "argocd"
    }
    spec = {
      selector = {
        matchLabels = {
          "app.kubernetes.io/name" = "argocd-repo-server"
        }
      }
      endpoints = [
        {
          port     = 8084
          path     = "/metrics"
          interval = "30s"
        }
      ]
    }
  }
  depends_on = [helm_release.argocd]
}

resource "kubernetes_manifest" "argocd_application_controller_monitoring" {
  manifest = {
    apiVersion = "monitoring.googleapis.com/v1"
    kind       = "PodMonitoring"
    metadata = {
      name      = "argocd-application-controller-monitoring"
      namespace = "argocd"
    }
    spec = {
      selector = {
        matchLabels = {
          "app.kubernetes.io/name" = "argocd-application-controller"
        }
      }
      endpoints = [
        {
          port     = 8082
          path     = "/metrics"
          interval = "30s"
        }
      ]
    }
  }
  depends_on = [helm_release.argocd]
}
