from datetime import datetime, timezone
from typing import Optional
from uuid import UUID

from sqlalchemy.orm import Session

from app.models.progress_models import DimUser
from app.schemas.dim_user_schema import DimUserCreate, DimUserUpdate


class DimUserService:
    def __init__(self, db: Session):
        self.db = db

    def get_user_by_id(self, user_id: UUID) -> Optional[DimUser]:
        return (
            self.db.query(DimUser)
            .filter(DimUser.user_id == user_id)
            .one_or_none()
        )

    def create_user(self, user_data: DimUserCreate) -> DimUser:
        user_dict = user_data.model_dump()
        if user_dict.get("locale") is None:
            user_dict["locale"] = "en"

        new_user = DimUser(**user_dict)
        self.db.add(new_user)
        self.db.commit()
        self.db.refresh(new_user)
        return new_user

    def update_user(self, user_id: UUID, user_data: DimUserUpdate) -> Optional[DimUser]:
        user = self.get_user_by_id(user_id)
        if not user:
            return None

        update_data = user_data.model_dump(exclude_unset=True)
        for field, value in update_data.items():
            setattr(user, field, value)

        user.updated_at = datetime.now(timezone.utc)
        self.db.commit()
        self.db.refresh(user)
        return user

    def update_locale(self, user_id: UUID, locale: str) -> Optional[DimUser]:
        return self.update_user(user_id, DimUserUpdate(locale=locale))

    def delete_user(self, user_id: UUID) -> bool:
        user = self.get_user_by_id(user_id)
        if not user:
            return False

        self.db.delete(user)
        self.db.commit()
        return True

    def user_exists(self, user_id: UUID) -> bool:
        return self.db.query(DimUser).filter(DimUser.user_id == user_id).count() > 0
