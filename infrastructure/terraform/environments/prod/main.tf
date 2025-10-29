locals {
  name       = "english-app-prod"
  aws_region = var.aws_region
  azs        = ["us-east-1a", "us-east-1b"]
  tags = {
    Environment = "prod"
    Project     = var.project_name
  }
}

module "network" {
  source = "../../modules/network"

  name               = local.name
  vpc_cidr           = "10.20.0.0/16"
  azs                = local.azs
  private_subnets    = ["10.20.1.0/24", "10.20.2.0/24"]
  public_subnets     = ["10.20.11.0/24", "10.20.12.0/24"]
  enable_nat_gateway = true
  single_nat_gateway = true
  tags               = local.tags
}

module "eks" {
  source = "../../modules/eks"

  cluster_name        = local.name
  vpc_id              = module.network.vpc_id
  private_subnets     = module.network.private_subnets
  public_subnets      = module.network.public_subnets
  use_private_subnets = true
  region              = local.aws_region

  node_groups = {
    on_demand_medium = {
      instance_types = ["t3.medium"]
      capacity_type  = "ON_DEMAND"
      desired_size   = 2
      min_size       = 2
      max_size       = 4
    }
    spot_small = {
      instance_types = ["t3.small"]
      capacity_type  = "SPOT"
      desired_size   = 1
      min_size       = 0
      max_size       = 3
    }
  }

  tags = local.tags
}

module "ecr" {
  source = "../../modules/ecr"

  repository_names = [
    "user-services",
    "lesson-services",
    "content-services",
    "notification-services",
    "bff-services",
  ]
}

resource "random_password" "db_password" {
  length  = 24
  special = true
}

resource "random_password" "redis_password" {
  length  = 24
  special = false
}

resource "random_password" "mq_password" {
  length  = 24
  special = true
}

module "secrets" {
  source = "../../modules/secrets"

  secrets = {
    "${local.name}/db/password"    = random_password.db_password.result
    "${local.name}/redis/password" = random_password.redis_password.result
    "${local.name}/mq/password"    = random_password.mq_password.result
  }

  tags = local.tags
}

module "rds" {
  source = "../../modules/rds-postgres"

  name                  = "${local.name}-postgres"
  engine_version        = "15"
  instance_class        = "db.t4g.small"
  vpc_id                = module.network.vpc_id
  subnet_ids            = module.network.private_subnets
  allowed_cidrs         = ["10.20.0.0/16"]
  multi_az              = true
  username              = "user"
  password              = random_password.db_password.result
  deletion_protection   = true
  allocated_storage     = 50
  max_allocated_storage = 200
  tags                  = local.tags
}

module "redis" {
  source = "../../modules/elasticache-redis"

  name              = "${local.name}-redis"
  node_type         = "cache.t4g.small"
  num_cache_nodes   = 2
  multi_az_enabled  = true
  vpc_id            = module.network.vpc_id
  subnet_ids        = module.network.private_subnets
  allowed_cidrs     = ["10.20.0.0/16"]
  tags              = local.tags
}

resource "aws_security_group" "mq_sg" {
  name   = "${local.name}-mq-sg"
  vpc_id = module.network.vpc_id

  ingress {
    from_port   = 5671
    to_port     = 5672
    protocol    = "tcp"
    cidr_blocks = ["10.20.0.0/16"]
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags = local.tags
}

module "mq" {
  source = "../../modules/amazon-mq-rabbit"

  name               = "${local.name}-mq"
  host_instance_type = "mq.t3.small"
  deployment_mode    = "ACTIVE_STANDBY_MULTI_AZ"
  username           = "user"
  password           = random_password.mq_password.result
  subnet_ids         = module.network.private_subnets
  security_groups    = [aws_security_group.mq_sg.id]
}

output "cluster_name" { value = module.eks.cluster_name }
output "vpc_id" { value = module.network.vpc_id }

