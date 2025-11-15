import logging
import os
from pathlib import Path

from alembic import command
from alembic.config import Config

from app.config import settings

logger = logging.getLogger(__name__)

BASE_DIR = Path(__file__).resolve().parents[2]


def load_alembic_config() -> Config:
    """Build an Alembic Config object tied to this project."""
    ini_name = os.getenv("ALEMBIC_INI_PATH", "alembic.ini")
    ini_path = (BASE_DIR / ini_name).resolve()

    if not ini_path.exists():
        raise FileNotFoundError(f"Alembic config not found at {ini_path}")

    cfg = Config(str(ini_path))
    script_location = os.getenv("ALEMBIC_SCRIPT_LOCATION", str(BASE_DIR / "alembic"))
    cfg.set_main_option("script_location", script_location)
    cfg.set_main_option("sqlalchemy.url", settings.database_url)
    return cfg


def run_database_migrations(target: str = "head") -> None:
    """Apply Alembic migrations up to the provided target (defaults to head)."""
    cfg = load_alembic_config()
    try:
        command.upgrade(cfg, target)
        logger.info("Database migrations applied (%s)", target)
    except Exception:
        logger.exception("Database migration failed")
        raise
