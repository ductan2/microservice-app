variable "name" { type = string }
variable "engine_version" { type = string default = "15" }
variable "instance_class" { type = string }
variable "allocated_storage" { type = number default = 20 }
variable "max_allocated_storage" { type = number default = 100 }
variable "multi_az" { type = bool default = false }
variable "vpc_id" { type = string }
variable "subnet_ids" { type = list(string) }
variable "allowed_cidrs" { type = list(string) }
variable "username" { type = string default = "user" }
variable "password" { type = string }
variable "deletion_protection" { type = bool default = false }
variable "tags" { type = map(string) default = {} }

