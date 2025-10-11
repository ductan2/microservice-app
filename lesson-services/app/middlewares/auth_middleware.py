from fastapi import Request, HTTPException
from uuid import UUID
from starlette.middleware.base import BaseHTTPMiddleware


class InternalAuthRequired(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        # Skip auth for health check endpoint
        if request.url.path == "/api/v1/health":
            response = await call_next(request)
            return response

        user_id = request.headers.get("X-User-ID")
        email = request.headers.get("X-User-Email")
        session_id = request.headers.get("X-Session-ID")

        if not user_id or not email or not session_id:
            raise HTTPException(status_code=401, detail="missing internal auth headers")

        try:
            parsed_user_id = UUID(user_id)
        except ValueError:
            raise HTTPException(status_code=401, detail="invalid user ID format")

        try:
            parsed_session_id = UUID(session_id)
        except ValueError:
            raise HTTPException(status_code=401, detail="invalid session ID format")

        request.state.user_id = parsed_user_id
        request.state.user_email = email
        request.state.session_id = parsed_session_id

        response = await call_next(request)
        return response
