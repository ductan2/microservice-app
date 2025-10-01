import os
from typing import Optional
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    database_url: str = "postgresql://user:password@localhost:5432/lesson_db"
    redis_url: str = "redis://localhost:6379"
    secret_key: str = "your-secret-key-change-in-production"
    algorithm: str = "HS256"
    access_token_expire_minutes: int = 30
    environment: str = "development"
    
    class Config:
        env_file = ".env"

settings = Settings()