"""Shared routing utilities."""

from __future__ import annotations

import json
from typing import Any, Dict

from fastapi import HTTPException, Request, status
from fastapi.responses import JSONResponse, Response
from fastapi.routing import APIRoute

from app.config import DEFAULT_SUCCESS_DATA, get_error_message


class ApiResponseRoute(APIRoute):
    """APIRoute that normalizes successful and error responses."""

    def get_route_handler(self):
        original_route_handler = super().get_route_handler()

        async def custom_route_handler(request: Request) -> Response:
            try:
                response: Response = await original_route_handler(request)
            except HTTPException as exc:
                return self._format_error_response(exc)
            except Exception as exc:  # noqa: BLE001 - bubble as normalized error payload
                return JSONResponse(
                    status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                    content={
                        "status": "error",
                        "message": get_error_message(status.HTTP_500_INTERNAL_SERVER_ERROR),
                        "error": str(exc),
                    },
                )

            if not isinstance(response, Response):
                return self._build_success_response(status.HTTP_200_OK, response)

            if self._should_passthrough(response):
                return response

            payload, already_formatted = self._parse_body(response)

            if response.status_code >= 400:
                return self._build_error_response(
                    response,
                    payload,
                    already_formatted=already_formatted,
                )

            if already_formatted:
                return JSONResponse(
                    status_code=response.status_code,
                    content=payload,
                    headers=self._copy_headers(response.headers),
                    background=response.background,
                )

            success_status = (
                response.status_code
                if response.status_code != status.HTTP_204_NO_CONTENT
                else status.HTTP_200_OK
            )

            return self._build_success_response(
                success_status,
                payload,
                headers=response.headers,
                background=response.background,
            )

        return custom_route_handler

    def _format_error_response(self, exc: HTTPException) -> JSONResponse:
        error_detail = self._normalize_error_detail(exc.detail)
        return JSONResponse(
            status_code=exc.status_code,
            content={
                "status": "error",
                "message": get_error_message(exc.status_code),
                "error": error_detail,
            },
            headers=exc.headers,
        )

    def _normalize_error_detail(self, detail: Any) -> Any:
        if isinstance(detail, dict):
            if "detail" in detail and len(detail) == 1:
                return detail["detail"]
            return detail
        if isinstance(detail, list):
            return detail
        return str(detail) if detail is not None else ""

    def _should_passthrough(self, response: Response) -> bool:
        media_type = getattr(response, "media_type", None)
        return media_type is not None and media_type != "application/json"

    def _parse_body(self, response: Response) -> tuple[Any, bool]:
        body = getattr(response, "body", b"")
        if not body:
            return DEFAULT_SUCCESS_DATA.copy(), False

        try:
            parsed = json.loads(body)
        except json.JSONDecodeError:
            if isinstance(body, (bytes, bytearray)):
                return body.decode(), False
            return body, False

        if isinstance(parsed, dict) and parsed.get("status") in {"success", "error"}:
            return parsed, True

        return parsed, False

    def _build_error_response(
        self,
        response: Response,
        payload: Any,
        *,
        already_formatted: bool,
    ) -> JSONResponse:
        cleaned_headers = self._copy_headers(response.headers)
        if already_formatted and isinstance(payload, dict):
            content = payload.copy()
            content.setdefault("status", "error")
            content.setdefault("message", get_error_message(response.status_code))
            return JSONResponse(
                status_code=response.status_code,
                content=content,
                headers=cleaned_headers,
                background=response.background,
            )

        error_detail = self._normalize_error_detail(payload)
        return JSONResponse(
            status_code=response.status_code,
            content={
                "status": "error",
                "message": get_error_message(response.status_code),
                "error": error_detail,
            },
            headers=cleaned_headers,
            background=response.background,
        )

    def _build_success_response(
        self,
        status_code: int,
        data: Any,
        *,
        headers: Dict[str, str] | None = None,
        background: Any | None = None,
    ) -> JSONResponse:
        normalized = self._normalize_success_data(data)
        cleaned_headers = self._copy_headers(headers)
        return JSONResponse(
            status_code=status_code,
            content={"status": "success", "data": normalized},
            headers=cleaned_headers,
            background=background,
        )

    def _normalize_success_data(self, data: Any) -> Any:
        if data is None or data == "" or data == DEFAULT_SUCCESS_DATA:
            return DEFAULT_SUCCESS_DATA.copy()
        return data

    def _copy_headers(self, headers: Dict[str, str] | None) -> Dict[str, str] | None:
        if not headers:
            return None
        excluded = {"content-length", "content-type"}
        return {key: value for key, value in headers.items() if key.lower() not in excluded}
