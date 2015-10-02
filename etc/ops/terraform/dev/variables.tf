variable "gcloud_project" {}
variable "gcloud_region" {}
variable "gcloud_account_file" {}
variable "gcloud_zone" {}
variable "gcloud_user" {}
variable "gcloud_image" {
  description = "Base image of the GCE Virtual Machine"
  default = "ubuntu-1504-vivid-v20150911"
}
variable "gcloud_machine_type" {
  default = "n1-standard-4"
}
variable "gcloud_disk_size" {
  description = "Size of the persistent SSD disk in gigabytes"
  default = "50"
}
