"""
main.py — FastAPI application for the Legal Metrology Rule Engine microservice.

Exposes:
  POST /validate  — Run compliance check on extracted label data
  GET  /rules     — List all active rules and ruleset version
  GET  /health    — Health check for load balancer / orchestration

This service does NOT perform OCR or ML. It receives already-extracted structured
label data and returns a structured violation report with legal citations.
"""

from __future__ import annotations

import json
from pathlib import Path

from fastapi import FastAPI, Request, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from engine import validate
from rule_registry import RULESET_DESCRIPTION, RULESET_VERSION, get_all_rule_metadata
from schemas import (
    RuleInfo,
    RuleRegistryResponse,
    ValidationRequest,
    ValidationResponse,
)

# ---------------------------------------------------------------------------
# App initialisation
# ---------------------------------------------------------------------------

app = FastAPI(
    title="Legal Metrology Rule Engine",
    description=(
        "Deterministic, rule-based compliance checking for packaged commodities "
        "under India's Legal Metrology (Packaged Commodities) Rules, 2011 "
        "and 2017 Amendments. Every violation cites its exact legal clause."
    ),
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
    contact={
        "name": "Legal Metrology Compliance System",
        "email": "compliance@example.gov.in",
    },
    license_info={
        "name": "Government of India — Ministry of Consumer Affairs",
    },
)

# ---------------------------------------------------------------------------
# CORS — allow all origins for dev/demo; restrict in production
# ---------------------------------------------------------------------------
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# ---------------------------------------------------------------------------
# Exception handler for clean error responses
# ---------------------------------------------------------------------------

@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception) -> JSONResponse:
    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content={
            "error": "Internal server error",
            "detail": str(exc),
            "path": str(request.url),
        },
    )


# ---------------------------------------------------------------------------
# Routes
# ---------------------------------------------------------------------------

@app.post(
    "/validate",
    response_model=ValidationResponse,
    summary="Validate product label compliance",
    description=(
        "Accepts extracted label fields as structured JSON and returns a compliance "
        "report with violations (each citing its exact legal clause), manual review "
        "flags, and overall compliance status. No OCR or ML is performed here."
    ),
    tags=["Compliance"],
    status_code=status.HTTP_200_OK,
)
async def validate_product(request: ValidationRequest) -> ValidationResponse:
    """
    Run the complete compliance rule engine against extracted label data.

    The engine:
    1. Checks bulk-package exemption (>25 kg / >50 kg for special categories)
    2. Determines which rules apply based on category, channel, and import status
    3. Filters low-confidence fields to manual_review_required
    4. Runs all applicable deterministic rules
    5. Returns structured violations with Rule citations
    """
    return validate(request)


@app.get(
    "/rules",
    response_model=RuleRegistryResponse,
    summary="List all active compliance rules",
    description=(
        "Returns the current active rule registry with legal citations, severity "
        "levels, and the ruleset version string. Useful for admin dashboards and "
        "audit trails."
    ),
    tags=["Administration"],
    status_code=status.HTTP_200_OK,
)
async def list_rules() -> RuleRegistryResponse:
    """
    Return metadata for all currently registered and active rules.

    The ruleset_version field indicates exactly which version of the
    Legal Metrology (Packaged Commodities) Rules is being enforced.
    """
    raw_metadata = get_all_rule_metadata()
    rules = [
        RuleInfo(
            rule_id=r["rule_id"],
            clause=r["clause"],
            severity=r["severity"],
            description=r["description"],
            rule_class=r["rule_class"],
        )
        for r in raw_metadata
    ]
    return RuleRegistryResponse(
        ruleset_version=RULESET_VERSION,
        ruleset_description=RULESET_DESCRIPTION,
        rules=rules,
    )


@app.get(
    "/health",
    summary="Health check",
    description="Basic liveness probe for load balancers and orchestration.",
    tags=["Operational"],
    status_code=status.HTTP_200_OK,
)
async def health_check() -> dict:
    """Return service liveness and current ruleset version."""
    return {
        "status": "healthy",
        "service": "legal-metrology-rule-engine",
        "ruleset_version": RULESET_VERSION,
        "ruleset_description": RULESET_DESCRIPTION,
    }


@app.get(
    "/",
    summary="Root redirect",
    include_in_schema=False,
)
async def root() -> dict:
    return {
        "service": "legal-metrology-rule-engine",
        "version": "1.0.0",
        "docs": "/docs",
        "health": "/health",
        "rules": "/rules",
        "validate": "POST /validate",
    }
