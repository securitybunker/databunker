
variable "name_suffix" {
  type = string
  default = "databunker"
}

variable "aws_access_key" {
  type = string
  default = ""
}

variable "aws_secret_key" {
  type = string
  default = ""
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "eu-north-1"
}

variable "resource_tags" {
  description = "Tags to set for all resources"
  type        = map(string)
  default     = {
    project     = "databunker",
    environment = "dev"
  }
}

variable "ec2_eks_instance_type" {
  description = "AWS EC2 instance type for EKS nodes."
  type        = string
  default     = "t3.medium"
}

variable "ec2_rds_instance_type" {
  description = "AWS EC2 instance type for RDS nodes."
  type        = string
  default     = "db.t3.medium"
}

resource "random_password" "db_password" {
  length           = 16
  special          = false
}
