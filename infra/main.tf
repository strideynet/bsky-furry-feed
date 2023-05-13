provider "google" {
  project     = "bsky-furry-feed"
}

provider "google-beta" {
  project     = "bsky-furry-feed"
}

resource "google_sql_database_instance" "main_us_east" {
  database_version = "POSTGRES_14"
  name             = "main-us-east"
  region           = "us-east1"

  settings {
    availability_type = "REGIONAL"
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
  }
}

