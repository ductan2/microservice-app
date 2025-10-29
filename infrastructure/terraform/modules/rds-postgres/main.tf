resource "aws_db_subnet_group" "this" {
  name       = "${var.name}-subnets"
  subnet_ids = var.subnet_ids
  tags       = var.tags
}

resource "aws_security_group" "this" {
  name        = "${var.name}-sg"
  description = "Security group for RDS ${var.name}"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = var.allowed_cidrs
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}

resource "aws_db_instance" "this" {
  identifier                 = var.name
  engine                     = "postgres"
  engine_version             = var.engine_version
  instance_class             = var.instance_class
  db_subnet_group_name       = aws_db_subnet_group.this.name
  vpc_security_group_ids     = [aws_security_group.this.id]
  allocated_storage          = var.allocated_storage
  max_allocated_storage      = var.max_allocated_storage
  username                   = var.username
  password                   = var.password
  multi_az                   = var.multi_az
  publicly_accessible        = false
  storage_encrypted          = true
  deletion_protection        = var.deletion_protection
  skip_final_snapshot        = true
  apply_immediately          = true
  auto_minor_version_upgrade = true
  backup_retention_period    = var.multi_az ? 7 : 1

  tags = var.tags
}

output "endpoint" { value = aws_db_instance.this.address }
output "port" { value = aws_db_instance.this.port }

