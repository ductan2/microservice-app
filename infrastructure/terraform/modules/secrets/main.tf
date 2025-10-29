resource "aws_secretsmanager_secret" "this" {
  for_each = var.secrets
  name     = each.key
  tags     = var.tags
}

resource "aws_secretsmanager_secret_version" "this" {
  for_each      = var.secrets
  secret_id     = aws_secretsmanager_secret.this[each.key].id
  secret_string = each.value
}

output "secret_arns" {
  value = { for k, v in aws_secretsmanager_secret.this : k => v.arn }
}

