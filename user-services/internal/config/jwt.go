package config

// GetJWTConfig returns the JWT configuration from the global config
func GetJWTConfig() JWTConfig {
	return GetConfig().JWT
}
