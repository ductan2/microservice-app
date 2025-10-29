variable "repository_names" {
  type = list(string)
}

resource "aws_ecr_repository" "this" {
  for_each             = toset(var.repository_names)
  name                 = each.key
  image_tag_mutability = "MUTABLE"
  image_scanning_configuration {
    scan_on_push = true
  }
}

output "repository_urls" {
  value = { for k, v in aws_ecr_repository.this : k => v.repository_url }
}

