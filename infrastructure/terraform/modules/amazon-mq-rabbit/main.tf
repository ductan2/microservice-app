variable "name" { type = string }
variable "engine_version" { type = string default = "3.13.21" }
variable "host_instance_type" { type = string }
variable "deployment_mode" { type = string default = "SINGLE_INSTANCE" } # or ACTIVE_STANDBY_MULTI_AZ
variable "username" { type = string default = "user" }
variable "password" { type = string }
variable "subnet_ids" { type = list(string) }
variable "security_groups" { type = list(string) }

resource "aws_mq_broker" "this" {
  broker_name        = var.name
  engine_type        = "RabbitMQ"
  engine_version     = var.engine_version
  host_instance_type = var.host_instance_type
  deployment_mode    = var.deployment_mode

  user {
    username = var.username
    password = var.password
  }

  logs {
    general = true
  }

  subnet_ids       = var.subnet_ids
  security_groups  = var.security_groups
  publicly_accessible = false
  auto_minor_version_upgrade = true
}

output "amqp_endpoints" { value = aws_mq_broker.this.instances[*].endpoints }

