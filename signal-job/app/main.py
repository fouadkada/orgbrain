import time
import structlog

logger = structlog.get_logger()

if __name__ == "__main__":
    logger.info("signal-job started", service="signal-job")
    while True:  # local-dev only: keeps container alive; production uses Coolify cron
        time.sleep(60)
