from typing import List
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Response, status
from sqlalchemy.orm import Session

from app.database.connection import get_db
from app.schemas.progress_schema import (
    QuizAnswerCreate,
    QuizAnswerResponse,
    QuizAnswerSummary,
    QuizAnswerUpdate,
)
from app.services.quiz_answer_service import QuizAnswerService


router = APIRouter(prefix="/api/quiz-answers", tags=["Quiz Answers"])


def _get_service(db: Session) -> QuizAnswerService:
    return QuizAnswerService(db)


@router.get("/attempt/{attempt_id}", response_model=List[QuizAnswerResponse])
def get_attempt_answers(
    attempt_id: UUID,
    db: Session = Depends(get_db),
) -> List[QuizAnswerResponse]:
    service = _get_service(db)
    return service.get_attempt_answers(attempt_id)


@router.post("", response_model=QuizAnswerResponse, status_code=status.HTTP_201_CREATED)
def create_answer(
    payload: QuizAnswerCreate,
    db: Session = Depends(get_db),
) -> QuizAnswerResponse:
    service = _get_service(db)
    try:
        return service.create_answer(payload)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(exc)) from exc


@router.get("/{answer_id}", response_model=QuizAnswerResponse)
def get_answer(answer_id: UUID, db: Session = Depends(get_db)) -> QuizAnswerResponse:
    service = _get_service(db)
    answer = service.get_answer(answer_id)
    if not answer:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Quiz answer not found")
    return answer


@router.put("/{answer_id}", response_model=QuizAnswerResponse)
def update_answer(
    answer_id: UUID,
    payload: QuizAnswerUpdate,
    db: Session = Depends(get_db),
) -> QuizAnswerResponse:
    service = _get_service(db)
    try:
        answer = service.update_answer(answer_id, payload)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(exc)) from exc

    if not answer:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Quiz answer not found")
    return answer


@router.delete("/{answer_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_answer(answer_id: UUID, db: Session = Depends(get_db)) -> Response:
    service = _get_service(db)
    try:
        deleted = service.delete_answer(answer_id)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(exc)) from exc

    if not deleted:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Quiz answer not found")
    return Response(status_code=status.HTTP_204_NO_CONTENT)


@router.get("/attempt/{attempt_id}/summary", response_model=QuizAnswerSummary)
def get_answer_summary(
    attempt_id: UUID,
    db: Session = Depends(get_db),
) -> QuizAnswerSummary:
    service = _get_service(db)
    return service.get_answer_summary(attempt_id)

