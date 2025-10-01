from pydantic import BaseModel
from typing import Optional
from datetime import datetime

class LessonBase(BaseModel):
    title: str
    description: Optional[str] = None
    content: str
    level: str
    duration_minutes: int
    is_active: bool = True

class LessonCreate(LessonBase):
    pass

class LessonUpdate(BaseModel):
    title: Optional[str] = None
    description: Optional[str] = None
    content: Optional[str] = None
    level: Optional[str] = None
    duration_minutes: Optional[int] = None
    is_active: Optional[bool] = None

class LessonResponse(LessonBase):
    id: int
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True