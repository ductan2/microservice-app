import os
from typing import Optional
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    # PostgreSQL settings
    postgres_user: str = "user"
    postgres_password: str = "password"
    postgres_db: str = "english_app"
    postgres_port: int = 5432
    postgres_host: str = "localhost"
    postgres_exporter_port: int = 9187
    
    # Redis settings
    redis_port: int = 6379
    redis_password: str = "redis_password"
    redis_host: str = "localhost"
    redis_exporter_port: int = 9121
    
    # RabbitMQ settings (for future use)
    rabbitmq_user: str = "user"
    rabbitmq_password: str = "password"
    rabbitmq_vhost: str = "/"
    rabbitmq_port: int = 5672
    rabbitmq_mgmt_port: int = 15672
    rabbitmq_exporter_port: int = 9419
    
    # Application settings
    secret_key: str = "your-secret-key-change-in-production"
    algorithm: str = "HS256"
    access_token_expire_minutes: int = 30
    environment: str = "development"
    
    @property
    def database_url(self) -> str:
        print(f"postgresql://{self.postgres_user}:{self.postgres_password}@{self.postgres_host}:{self.postgres_port}/{self.postgres_db}")
        return f"postgresql://{self.postgres_user}:{self.postgres_password}@{self.postgres_host}:{self.postgres_port}/{self.postgres_db}"
    
    @property
    def redis_url(self) -> str:
        if self.redis_password:
            return f"redis://:{self.redis_password}@{self.redis_host}:{self.redis_port}"
        return f"redis://{self.redis_host}:{self.redis_port}"
    
    class Config:
        env_file = ".env"
        extra = "ignore"  # Ignore extra environment variables

settings = Settings()