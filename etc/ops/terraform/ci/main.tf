provider "google" {
    account_file = ""
    project = "${var.gcloud_project}"
    region = "${var.gcloud_region}"
}

resource "google_compute_address" "ci" {
  count = "${var.ci_count}"
  name = "ci${count.index}"
}

resource "google_compute_instance" "ci" {
  count = "${var.ci_count}"
  name = "ci${count.index}"
  machine_type = "${var.gcloud_machine_type}"
  zone = "${var.gcloud_zone}"
  tags = ["ci"]
  disk {
    type = "pd-ssd"
    image = "${var.gcloud_image}"
  }
  metadata_startup_script = "${file("terraform/dev/startup.sh")}"
  metadata {
    agent_token = "${var.ci_buildkite_agent_token}"
  }
  network_interface {
    network = "default"
    access_config {
      nat_ip = "${element(google_compute_address.ci.*.address, count.index)}"
    }
  }
}
