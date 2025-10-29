variable "cluster_name" { type = string }
variable "vpc_id" { type = string }
variable "private_subnets" { type = list(string) }
variable "public_subnets" { type = list(string) }
variable "use_private_subnets" { type = bool }
variable "region" { type = string }
variable "node_groups" {
  description = "Map of managed node group configurations"
  type = any
}
variable "tags" { type = map(string) default = {} }

