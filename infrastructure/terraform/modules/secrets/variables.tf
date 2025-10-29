variable "secrets" {
  description = "Map of secret name to secret string value"
  type        = map(string)
}

variable "tags" { type = map(string) default = {} }

