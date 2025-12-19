# Binary Authorization Policy
resource "google_binary_authorization_policy" "policy" {
  admission_whitelist_patterns {
    name_pattern = "gcr.io/google_containers/*"
  }

  admission_whitelist_patterns {
    name_pattern = "gcr.io/google-containers/*"
  }

  admission_whitelist_patterns {
    name_pattern = "k8s.gcr.io/*"
  }

  admission_whitelist_patterns {
    name_pattern = "gke.gcr.io/*"
  }

  default_admission_rule {
    evaluation_mode  = "REQUIRE_ATTESTATION"
    enforcement_mode = "ENFORCED_BLOCK_AND_AUDIT_LOG"
    require_attestations_by = [
      google_binary_authorization_attestor.cosign_attestor.name
    ]
  }

  # Allow system images to run without attestation
  cluster_admission_rules {
    cluster                 = "${var.region}.${google_container_cluster.primary.name}"
    evaluation_mode         = "REQUIRE_ATTESTATION"
    enforcement_mode        = "ENFORCED_BLOCK_AND_AUDIT_LOG"
    require_attestations_by = [
      google_binary_authorization_attestor.cosign_attestor.name
    ]
  }
}

# Attestor for Cosign signatures
resource "google_binary_authorization_attestor" "cosign_attestor" {
  name = "cosign-attestor"
  attestation_authority_note {
    note_reference = google_container_analysis_note.cosign_note.name
  }
}

# Container Analysis Note for attestations
resource "google_container_analysis_note" "cosign_note" {
  name = "cosign-attestor-note"
  attestation_authority {
    hint {
      human_readable_name = "Cosign keyless signatures from GitHub Actions"
    }
  }
}
