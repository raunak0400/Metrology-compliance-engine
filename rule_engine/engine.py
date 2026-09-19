"""
engine.py — Orchestrator for the Legal Metrology Rule Engine.

This module is the main entry point for compliance validation. It:
  1. Determines which rules apply (via applicability.py)
  2. Checks for bulk exemption — returns early if exempt
  3. Splits fields into high-confidence and low-confidence buckets
  4. Runs each applicable rule against high-confidence fields
  5. Collects all violations and manual-review flags
  6. Assembles and returns a structured ValidationResponse

Design principle: This module contains NO compliance logic itself — all decisions
live in individual rule files. The engine is purely an orchestrator.
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import Dict, List, Tuple

from applicability import determine_applicable_rules, is_bulk_exempt
from rule_registry import (
    RULE_REGISTRY,
    RULESET_VERSION,
    MULTI_VIOLATION_RULES,
    get_all_rule_metadata,
)
from rules.font_size_rule import FontSizeRule
from rules.placement_rule import PlacementRule
from schemas import (
    ApplicabilityNote,
    ExtractedField,
    ValidationRequest,
    ValidationResponse,
    ViolationResult,
)
from tag_bridge import synthesize_from_tags

# Load confidence threshold from config — single source of truth
_CONFIG_PATH = Path(__file__).parent / "rules_config.json"
with _CONFIG_PATH.open(encoding="utf-8") as _f:
    _CONFIG = json.load(_f)

CONFIDENCE_THRESHOLD: float = _CONFIG["confidence_threshold"]


def _split_by_confidence(
    fields: List[ExtractedField],
) -> Tuple[List[ExtractedField], List[str]]:
    """
    Split extracted fields into high-confidence and low-confidence groups.

    A field_type only lands in manual_review_required if NONE of its
    extractions met the confidence threshold.  If a high-confidence
    extraction exists for a type, that type is excluded from manual_review
    even if a lower-confidence duplicate also exists for the same type.

    Returns:
        (high_confidence_fields, low_confidence_field_types)
    """
    high: List[ExtractedField] = []
    low_types_seen: set = set()

    for field in fields:
        if field.confidence >= CONFIDENCE_THRESHOLD:
            high.append(field)
        else:
            low_types_seen.add(field.field_type)

    # Only flag as manual-review if no high-confidence extraction covers that type
    high_types = {f.field_type for f in high}
    low_types = [t for t in low_types_seen if t not in high_types]

    return high, low_types


def validate(request: ValidationRequest) -> ValidationResponse:
    """
    Run a full compliance check for the given ValidationRequest.

    Pipeline:
      Step 0: Tag bridge — synthesize ExtractedField objects from extracted_tags
              for any field_type not already in extracted_fields.
      Step 1: Bulk exemption — return immediately if package exceeds threshold.
      Step 2: Determine applicable rules (category, channel, import status).
      Step 3: Confidence split — separate high/low confidence fields.
      Step 4: Run each applicable rule.
      Step 5: Determine overall_status.
      Step 6: Assemble and return ValidationResponse.
    """
    # -------------------------------------------------------------------------
    # Step 0: Tag bridge — synthesize missing fields from extracted_tags
    # -------------------------------------------------------------------------
    # The upstream OCR service may send structured extracted_tags alongside
    # (or instead of) extracted_fields. For every mandatory field_type absent
    # from extracted_fields, the bridge builds a synthetic ExtractedField from
    # the corresponding tag values. Caller-provided fields always win.
    synthesized: list[ExtractedField] = []
    synthesis_summaries: list[dict] = []

    if request.extracted_tags:
        synthesized, synthesis_summaries = synthesize_from_tags(
            tags=request.extracted_tags,
            measurements=request.text_measurements,
            existing_fields=request.extracted_fields,
            confidence_threshold=CONFIDENCE_THRESHOLD,
        )

    # Merge: caller-provided fields first, then synthesized gap-fillers
    all_fields: list[ExtractedField] = list(request.extracted_fields) + synthesized

    # -------------------------------------------------------------------------
    # Step 1: Check bulk exemption — if exempt, return immediately
    # -------------------------------------------------------------------------
    exempt, exempt_reason = is_bulk_exempt(request)
    if exempt:
        return ValidationResponse(
            product_id=request.product_id,
            ruleset_version=RULESET_VERSION,
            overall_status="exempt",
            exempt_reason=exempt_reason,
            violations=[],
            manual_review_required=[],
            applicability_notes=[
                ApplicabilityNote(
                    rule_id="ALL",
                    applied=False,
                    reason=exempt_reason,
                )
            ],
        )

    # -------------------------------------------------------------------------
    # Step 2: Determine which rules apply
    # -------------------------------------------------------------------------
    applicability = determine_applicable_rules(request)

    # Build applicability notes for response transparency
    applicability_notes: List[ApplicabilityNote] = []
    for rule_key, rule_config in applicability.items():
        rule_instance = RULE_REGISTRY.get(rule_key)
        rule_id = rule_instance.rule_id if rule_instance else rule_key.upper()
        applicability_notes.append(
            ApplicabilityNote(
                rule_id=rule_id,
                applied=rule_config["apply"],
                reason=rule_config["reason"],
                severity_override=rule_config.get("severity_override"),
            )
        )

    # -------------------------------------------------------------------------
    # Step 3: Split fields by confidence
    # -------------------------------------------------------------------------
    # all_fields = caller-provided extracted_fields + tag-synthesized fields
    high_confidence_fields, manual_review_types = _split_by_confidence(all_fields)

    # -------------------------------------------------------------------------
    # Step 4: Run applicable rules
    # -------------------------------------------------------------------------
    violations: List[ViolationResult] = []

    # Build set of field_types whose ONLY extractions were low-confidence.
    # These field types cannot be judged: running a rule against them would
    # incorrectly fire a "field missing" violation when the field IS present
    # but untrustworthy.  The rule_key matches the field_type directly for
    # all single-field rules (manufacturer, mrp, mfg_date, consumer_care, etc.).
    low_confidence_type_set = set(manual_review_types)
    # Build set of field_types that DO have at least one high-confidence extraction
    high_confidence_type_set = {f.field_type for f in high_confidence_fields}

    for rule_key, rule_config in applicability.items():
        if not rule_config["apply"]:
            continue  # Applicability says skip this rule

        # Single-field rules: if every extraction of that field_type was
        # low-confidence (i.e., the type is in manual_review but NOT in
        # high_confidence_type_set), skip the rule entirely so we don't
        # produce a spurious "field missing" violation.
        if (
            rule_key in low_confidence_type_set
            and rule_key not in high_confidence_type_set
            and rule_key not in MULTI_VIOLATION_RULES
        ):
            continue

        rule_instance = RULE_REGISTRY.get(rule_key)
        if rule_instance is None:
            continue  # Rule key registered in applicability but not in registry — skip

        # Determine effective severity (may be overridden by applicability)
        severity_override = rule_config.get("severity_override")

        if rule_key in MULTI_VIOLATION_RULES:
            # Font size and placement rules can return multiple violations
            if isinstance(rule_instance, FontSizeRule):
                rule_violations = rule_instance.check_all(
                    high_confidence_fields, request
                )
            elif isinstance(rule_instance, PlacementRule):
                rule_violations = rule_instance.check_all(
                    high_confidence_fields, request
                )
            else:
                single = rule_instance.check(high_confidence_fields, request)
                rule_violations = [single] if single else []
        else:
            single = rule_instance.check(high_confidence_fields, request)
            rule_violations = [single] if single else []

        for v in rule_violations:
            if severity_override:
                # Create a modified copy with overridden severity
                v = ViolationResult(
                    rule_id=v.rule_id,
                    clause=v.clause,
                    severity=severity_override,
                    description=f"[{severity_override.upper()}] {v.description}",
                    field_evidence=v.field_evidence,
                    field_type=v.field_type,
                )
            violations.append(v)

    # -------------------------------------------------------------------------
    # Step 5: Determine overall status
    # -------------------------------------------------------------------------
    if violations:
        overall_status = "non_compliant"
    elif manual_review_types:
        overall_status = "requires_manual_review"
    else:
        overall_status = "compliant"

    # -------------------------------------------------------------------------
    # Step 6: Assemble and return response
    # -------------------------------------------------------------------------
    return ValidationResponse(
        product_id=request.product_id,
        ruleset_version=RULESET_VERSION,
        overall_status=overall_status,
        exempt_reason=None,
        violations=violations,
        manual_review_required=manual_review_types,
        applicability_notes=applicability_notes,
        synthesized_fields_summary=synthesis_summaries,
    )
