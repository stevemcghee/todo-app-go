resource "helm_release" "gatekeeper" {
  name       = "gatekeeper"
  repository = "https://open-policy-agent.github.io/gatekeeper/charts"
  chart      = "gatekeeper"
  namespace  = "gatekeeper-system"
  create_namespace = true
  version    = "3.14.0" # Recent stable version

  set {
    name  = "replicas"
    value = "2"
  }
}

resource "helm_release" "gatekeeper_secondary" {
  provider = helm.secondary
  name       = "gatekeeper"
  repository = "https://open-policy-agent.github.io/gatekeeper/charts"
  chart      = "gatekeeper"
  namespace  = "gatekeeper-system"
  create_namespace = true
  version    = "3.14.0"

  set {
    name  = "replicas"
    value = "2"
  }
}
