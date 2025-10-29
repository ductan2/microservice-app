module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name    = var.cluster_name
  cluster_version = "1.29"

  vpc_id     = var.vpc_id
  subnet_ids = var.use_private_subnets ? var.private_subnets : var.public_subnets

  eks_managed_node_groups = var.node_groups

  tags = var.tags
}

output "cluster_name" { value = module.eks.cluster_name }
output "cluster_endpoint" { value = module.eks.cluster_endpoint }
output "cluster_certificate_authority_data" { value = module.eks.cluster_certificate_authority_data }
output "cluster_security_group_id" { value = module.eks.cluster_security_group_id }
output "oidc_provider_arn" { value = module.eks.oidc_provider_arn }

