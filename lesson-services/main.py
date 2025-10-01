from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.routers import lesson_routes, user_routes, health_routes, progress_routes

app = FastAPI(
    title="Lesson Services API",
    description="A RESTful API for managing English learning lessons",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(health_routes.router, prefix="/api/v1", tags=["health"])
app.include_router(lesson_routes.router, prefix="/api/v1", tags=["lessons"])
app.include_router(user_routes.router, prefix="/api/v1", tags=["users"])
app.include_router(progress_routes.router, prefix="/api/v1/progress", tags=["progress"])

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)