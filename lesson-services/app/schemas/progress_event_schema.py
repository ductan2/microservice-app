from pydantic import BaseModel
from datetime import datetime
from uuid import UUID


# Progress Event Schemas
class ProgressEventBase(BaseModel):
    user_id: UUID
    type: str
    payload: dict


class ProgressEventCreate(ProgressEventBase):
    pass


class ProgressEventResponse(ProgressEventBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True
