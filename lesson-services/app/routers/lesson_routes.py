from fastapi import APIRouter, HTTPException, status
from typing import List, Optional
from app.schemas.lesson_schema import LessonCreate, LessonUpdate, LessonResponse
from app.services.lesson_service import LessonService

router = APIRouter()
lesson_service = LessonService()

@router.post("/lessons", status_code=status.HTTP_201_CREATED)
async def create_lesson(lesson: LessonCreate) -> LessonResponse:
    return await lesson_service.create_lesson(lesson)

@router.get("/lessons", status_code=status.HTTP_200_OK)
async def get_lessons(
    skip: int = 0,
    limit: int = 100,
    level: Optional[str] = None
) -> List[LessonResponse]:
    return await lesson_service.get_lessons(skip=skip, limit=limit, level=level)

@router.get("/lessons/{lesson_id}", status_code=status.HTTP_200_OK)
async def get_lesson(lesson_id: int) -> LessonResponse:
    lesson = await lesson_service.get_lesson(lesson_id)
    if not lesson:
        raise HTTPException(status_code=404, detail="Lesson not found")
    return lesson

@router.put("/lessons/{lesson_id}", status_code=status.HTTP_200_OK)
async def update_lesson(lesson_id: int, lesson: LessonUpdate) -> LessonResponse:
    updated_lesson = await lesson_service.update_lesson(lesson_id, lesson)
    if not updated_lesson:
        raise HTTPException(status_code=404, detail="Lesson not found")
    return updated_lesson

@router.delete("/lessons/{lesson_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_lesson(lesson_id: int):
    success = await lesson_service.delete_lesson(lesson_id)
    if not success:
        raise HTTPException(status_code=404, detail="Lesson not found")