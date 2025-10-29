variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "microservice-english-app"
}

variable "aws_region" {
  description = "AWS region to deploy the bootstrap resources"
  type        = string
  default     = "us-east-1"
}

variable "state_bucket_name" {
  description = "Name of the S3 bucket to store Terraform states"
  type        = string
  default     = "microservice-english-app-tfstate"
}

variable "state_lock_table_name" {
  description = "Name of the DynamoDB table for Terraform state locking"
  type        = string
  default     = "microservice-english-app-terraform-locks"
}


