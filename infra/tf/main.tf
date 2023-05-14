provider "google" {
  project     = "bsky-furry-feed"
}

provider "google-beta" {
  project     = "bsky-furry-feed"
}

data "google_compute_default_service_account" "default" {
}

resource "google_sql_database_instance" "main_us_east" {
  database_version = "POSTGRES_14"
  name             = "main-us-east"
  region           = "us-east1"

  settings {
    availability_type = "ZONAL"
    disk_autoresize   = true
    disk_size         = 15
    disk_type         = "PD_SSD"
    tier              = "db-f1-micro"
    deletion_protection_enabled = true

    backup_configuration {
      enabled            = true
      start_time         = "04:00"
      point_in_time_recovery_enabled = true
    }

    ip_configuration {
      ipv4_enabled = true
    }

    database_flags {
      name  = "cloudsql.iam_authentication"
      value = "on"
    }

    insights_config {
      query_insights_enabled = true
    }
  }
}

resource "google_sql_database" "database" {
  name     = "bff"
  instance = google_sql_database_instance.main_us_east.name
}

resource "google_sql_user" "main_us_east_default_compute_service_account" {
  name = replace(data.google_compute_default_service_account.default.email, ".gserviceaccount.com", "")
  instance = google_sql_database_instance.main_us_east.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
}

resource "google_sql_user" "main_us_east_noah" {
  name     = "noah@noahstride.co.uk"
  instance = google_sql_database_instance.main_us_east.name
  type     = "CLOUD_IAM_USER"
}

resource "google_container_cluster" "us_east" {
  name               = "us-east"
  location           = "us-east1"
  enable_autopilot = true
  ip_allocation_policy {
    cluster_ipv4_cidr_block  = ""
    services_ipv4_cidr_block = ""
  }
}

resource "google_service_account_iam_member" "bff_ingester_workload_identity_binding" {
  service_account_id = data.google_compute_default_service_account.default.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:bsky-furry-feed.svc.id.goog[default/bff-ingester]"
}

resource "google_compute_global_address" "ingress" {
  name         = "ingress"
  address_type = "EXTERNAL"
}

resource "google_dns_managed_zone" "furrylist" {
  name = "furrylist"
  dns_name = "furryli.st."
}

resource "google_dns_record_set" "feed_furrylist" {
  name         = "feed.${google_dns_managed_zone.furrylist.dns_name}"
  managed_zone = google_dns_managed_zone.furrylist.name
  type         = "A"
  ttl          = 300

  rrdatas = [google_compute_global_address.ingress.address]
}
