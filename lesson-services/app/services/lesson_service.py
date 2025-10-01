from typing import List, Optional
from app.schemas.lesson_schema import LessonCreate, LessonUpdate, LessonResponse

class LessonService:
    def __init__(self):
        pass
    
    async def create_lesson(self, lesson: LessonCreate) -> LessonResponse:
        pass
    
    async def get_lessons(self, skip: int = 0, limit: int = 100, level: Optional[str] = None) -> List[LessonResponse]:
        pass
    
    async def get_lesson(self, lesson_id: int) -> Optional[LessonResponse]:
        pass
    
    async def update_lesson(self, lesson_id: int, lesson: LessonUpdate) -> Optional[LessonResponse]:
        pass
    
    async def delete_lesson(self, lesson_id: int) -> bool:
        pass