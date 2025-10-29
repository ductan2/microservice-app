variable "name" { type = string }
variable "node_type" { type = string }
variable "num_cache_nodes" { type = number default = 1 }
variable "multi_az_enabled" { type = bool default = false }
variable "vpc_id" { type = string }
variable "subnet_ids" { type = list(string) }
variable "allowed_cidrs" { type = list(string) }
variable "tags" { type = map(string) default = {} }

resource "aws_elasticache_subnet_group" "this" {
  name       = "${var.name}-subnets"
  subnet_ids = var.subnet_ids
}

resource "aws_security_group" "this" {
  name   = "${var.name}-sg"
  vpc_id = var.vpc_id

  ingress {
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = var.allowed_cidrs
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "this" {
  replication_group_id          = var.name
  description                   = "Redis for ${var.name}"
  node_type                     = var.node_type
  number_cache_clusters         = var.num_cache_nodes
  automatic_failover_enabled    = var.multi_az_enabled
  multi_az_enabled              = var.multi_az_enabled
  subnet_group_name             = aws_elasticache_subnet_group.this.name
  security_group_ids            = [aws_security_group.this.id]
  port                          = 6379
  at_rest_encryption_enabled    = true
  transit_encryption_enabled    = true
  auto_minor_version_upgrade    = true
}

output "primary_endpoint_address" { value = aws_elasticache_replication_group.this.primary_endpoint_address }

