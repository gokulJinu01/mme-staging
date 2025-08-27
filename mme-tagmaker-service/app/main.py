from fastapi import FastAPI
from prometheus_fastapi_instrumentator import Instrumentator
from apscheduler.schedulers.background import BackgroundScheduler
from app.router import router
from app.services.tiering import rebalance_all_tags
from app.services.client import replay_failed_deltas
from app.services.database import db_service
from app.security.middleware import SecurityMiddleware, SecurityConfig
from app.security.handlers import security_router, set_security_middleware
from app.routes.edge_admin import router as edge_admin_router
from app.jobs.edge_learning import run_edge_learning

# Initialize security configuration
security_config = SecurityConfig()
security_middleware = SecurityMiddleware(None, security_config)

app = FastAPI(
    title="mme-tagmaker-service", 
    description="AI-powered tag extraction and tiering service for MME",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc"
)

# Add security middleware
app.add_middleware(SecurityMiddleware, config=security_config)

# Set up security handlers
set_security_middleware(security_middleware)

# Include routers
app.include_router(router)
app.include_router(security_router)
app.include_router(edge_admin_router)

Instrumentator().instrument(app).expose(app)

scheduler = BackgroundScheduler(timezone="UTC")
scheduler.add_job(rebalance_all_tags, "cron", hour=2, minute=0)  # Daily rebalancing at 2 AM
scheduler.add_job(replay_failed_deltas, "interval", minutes=1)   # Retry failed deltas every minute
scheduler.add_job(run_edge_learning, "interval", minutes=10)     # Edge learning every 10 minutes
scheduler.start()

@app.get("/health")
async def health():
    return {"status": "ok"}

@app.post("/feedback")
async def feedback():
    return {"message": "Feedback endpoint available", "status": "ok"}

@app.get("/version")
async def version():
    return {"version": "1.0.0", "service": "mme-tagmaker-service", "status": "ok"}

@app.on_event("shutdown")
async def shutdown_event():
    """Clean up resources on shutdown"""
    scheduler.shutdown()
    db_service.close()
