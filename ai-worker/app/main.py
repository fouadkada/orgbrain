from fastapi import FastAPI
import structlog

logger = structlog.get_logger()
# All log calls: logger.info("msg", service="ai-worker", trace_id=..., org_id=..., duration_ms=...)

app = FastAPI(title="ai-worker", version="0.1.0")


@app.get("/healthz")
async def healthz():
    return {"status": "ok"}
