from datetime import datetime
from typing import Optional
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field


class DimUserBase(BaseModel):
    user_id: UUID
    locale: Optional[str] = Field(default=None, max_length=50)
    level_hint: Optional[str] = Field(default=None, max_length=50)


class DimUserCreate(DimUserBase):
    locale: Optional[str] = Field(default="en", max_length=50)


class DimUserUpdate(BaseModel):
    locale: Optional[str] = Field(default=None, max_length=50)
    level_hint: Optional[str] = Field(default=None, max_length=50)


class DimUserLocaleUpdate(BaseModel):
    locale: str = Field(max_length=50)


class DimUserResponse(DimUserBase):
    updated_at: datetime

    model_config = ConfigDict(from_attributes=True)
