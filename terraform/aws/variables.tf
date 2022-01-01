variable "aws_access_key" {}
variable "aws_secret_key" {}
variable "region" {
default = "eu-north-1"
}

resource "random_password" "db_password" {
  length           = 16
  special          = true
  override_special = "@/"
}
