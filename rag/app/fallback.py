from enum import Enum

# FallbackRouter is a pure function — no side effects, no DB calls, no logging.
# All four outcomes must have unit tests (Story 1.4).


class FallbackOutcome(str, Enum):
    ROUTE_TO_OWNER = "ROUTE_TO_OWNER"
    NO_COVERAGE = "NO_COVERAGE"
    REPHRASE = "REPHRASE"
    ACCESS_FILTERED = "ACCESS_FILTERED"


# TODO: Implement route() pure function returning FallbackOutcome (Story 1.4).
