from fastapi import Request, HTTPException
from uuid import UUID
from starlette.middleware.base import BaseHTTPMiddleware


class InternalAuthRequired(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        # Skip auth for health check and other non-API endpoints
        skip_paths = [
            "/api/v1/health",
            "/favicon.ico",
            "/docs",
            "/redoc",
            "/openapi.json"
        ]
        
        if request.url.path in skip_paths or not request.url.path.startswith("/api/v1"):
            response = await call_next(request)
            return response

        # user_id = request.headers.get("X-User-ID")
        # email = request.headers.get("X-User-Email")
        # session_id = request.headers.get("X-Session-ID")

        # if not user_id or not email or not session_id:
        #     raise HTTPException(status_code=401, detail="missing internal auth headers")

        # try:
        #     parsed_user_id = UUID(user_id)
        # except ValueError:
        #     raise HTTPException(status_code=401, detail="invalid user ID format")

        # try:
        #     parsed_session_id = UUID(session_id)
        # except ValueError:
        #     raise HTTPException(status_code=401, detail="invalid session ID format")

    

        fake_user_id = UUID("612339a4-5b05-42f6-99e3-92b802044699")
        fake_email = "nygisagu@forexzig.com"
        fake_session_id = UUID("58702f3b-cdc5-481c-aae9-7bb02e096ad7")
        request.state.user_id = fake_user_id
        request.state.user_email = fake_email
        request.state.session_id = fake_session_id

        response = await call_next(request)
        return response
