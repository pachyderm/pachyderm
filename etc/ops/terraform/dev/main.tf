provider "google" {
    account_file = ""
    project = "${var.gcloud_project}"
    region = "${var.gcloud_region}"
}

resource "google_compute_address" "dev" {
  name = "dev-${var.gcloud_user}"
}

resource "google_compute_disk" "dev" {
  name = "dev-${var.gcloud_user}"
  type = "pd-ssd"
  zone = "${var.gcloud_zone}"
  size = "${var.gcloud_disk_size}"
  image = "${var.gcloud_image}"
}

resource "google_compute_instance" "dev" {
  name = "dev-${var.gcloud_user}"
  machine_type = "${var.gcloud_machine_type}"
  zone = "${google_compute_disk.dev.zone}"
  tags = ["dev"]
  disk {
    disk = "${google_compute_disk.dev.name}"
  }
  service_account {
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }
  metadata_startup_script = "${file("terraform/dev/startup.sh")}"
  network_interface {
    network = "default"
    access_config {
      nat_ip = "${google_compute_address.dev.address}"
    }
  }
}
