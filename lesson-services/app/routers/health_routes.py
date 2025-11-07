from typing import Dict

from fastapi import APIRouter, status

from app.routers.base import ApiResponseRoute

router = APIRouter(route_class=ApiResponseRoute)

@router.get("/health", status_code=status.HTTP_200_OK)
async def health_check() -> Dict[str, str]:
    return {"status": "healthy", "service": "lesson-services"}
