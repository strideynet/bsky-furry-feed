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
    disk_size         = 30
    disk_type         = "PD_HDD"
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
