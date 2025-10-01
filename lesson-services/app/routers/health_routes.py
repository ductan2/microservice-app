from fastapi import APIRouter, status
from typing import Dict

router = APIRouter()

@router.get("/health", status_code=status.HTTP_200_OK)
async def health_check() -> Dict[str, str]:
    return {"status": "healthy", "service": "lesson-services"}