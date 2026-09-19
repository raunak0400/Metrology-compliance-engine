"""
schemas.py — Pydantic models for the Legal Metrology Rule Engine.

Defines the input (ValidationRequest) and output (ValidationResponse) contracts
for the POST /validate endpoint.
"""

from __future__ import annotations

from typing import Any, Dict, List, Optional
from pydantic import BaseModel, ConfigDict, Field, field_validator


# ---------------------------------------------------------------------------
# Sub-models
# ---------------------------------------------------------------------------

class TextMeasurementEntry(BaseModel):
    """
    Per-text-block font height measurement from the upstream OCR service.
    Used by the tag bridge to attach font_height_mm to synthesized fields.
    """
    model_config = ConfigDict(extra="ignore")
    text: str
    confidence: float
    height_mm: float
    min_required_mm: float = 0.0
    is_rule7_compliant: bool = True


class BoundingBox(BaseModel):
    """Pixel/mm coordinates of a field on the label image."""

    x: float = Field(..., description="Left coordinate")
    y: float = Field(..., description="Top coordinate")
    w: float = Field(..., description="Width")
    h: float = Field(..., description="Height")


class ExtractedField(BaseModel):
    """
    A single field extracted from a product label by an upstream OCR/ML pipeline.
    The rule engine receives these and validates them — it never does any extraction.
    """

    field_type: str = Field(
        ...,
        description=(
            "Semantic label of the field. One of: manufacturer, generic_name, "
            "net_quantity, mrp, mfg_date, consumer_care, country_of_origin"
        ),
    )
    text: str = Field(..., description="Raw text of the extracted field")
    confidence: float = Field(
        ..., ge=0.0, le=1.0, description="Extraction confidence (0–1)"
    )
    bbox: BoundingBox = Field(..., description="Bounding box on the label")
    font_height_mm: Optional[float] = Field(
        None, ge=0.0, description="Estimated font height in millimetres"
    )
    panel_location: str = Field(
        ...,
        description=(
            "Which physical panel the field appears on. "
            "One of: principal_display_panel | other"
        ),
    )

    @field_validator("field_type")
    @classmethod
    def validate_field_type(cls, v: str) -> str:
        # Known mandatory field types used by the rule engine.
        # Unknown types (e.g. "unclassified_text" from extra_fields) are allowed
        # through unchanged — the engine ignores field_types it doesn't recognise.
        known = {
            "manufacturer",
            "generic_name",
            "net_quantity",
            "mrp",
            "mfg_date",
            "consumer_care",
            "country_of_origin",
        }
        # Pass-through unknown types (they land in extra_fields, not extracted_fields)
        return v

    @field_validator("panel_location")
    @classmethod
    def validate_panel_location(cls, v: str) -> str:
        allowed = {"principal_display_panel", "other"}
        if v not in allowed:
            raise ValueError(
                f"panel_location must be one of {sorted(allowed)}, got '{v}'"
            )
        return v


class ValidationRequest(BaseModel):
    """
    Input to the rule engine for a single product validation run.

    Only the fields declared here are used — everything else in the incoming
    JSON (panel_analysis, text_measurements, image_base64, extracted_tags, etc.)
    is silently dropped by Pydantic before the engine ever sees it.
    The upstream OCR service can send its full output envelope directly;
    the rule engine just picks out what it needs.
    """
    model_config = ConfigDict(extra="ignore")

    product_id: str = Field(..., description="Unique identifier for this product/batch")
    category: str = Field(
        ...,
        description="Product category: food | cosmetics | electronics | other | cement | fertilizer | agricultural_produce",
    )
    package_shape: str = Field(
        ...,
        description="Physical shape: rectangular | cylindrical | irregular (case-insensitive)",
    )
    package_surface_type: str = Field(
        ...,
        description="printed | embossed_molded_blown — determines Rule 7 font threshold",
    )
    sale_channel: str = Field(
        ...,
        description="retail | ecommerce — ecommerce skips the mfg_date rule per Rule 6(10)",
    )
    net_weight_kg: Optional[float] = Field(
        None,
        ge=0.0,
        description="Net weight in kg — triggers bulk exemption if > threshold for category",
    )
    net_volume_l: Optional[float] = Field(
        None,
        ge=0.0,
        description="Net volume in litres — triggers volume-based bulk exemption if > threshold",
    )
    is_imported: bool = Field(
        ...,
        description="True for imported goods — activates the country_of_origin rule",
    )
    extracted_fields: List[ExtractedField] = Field(
        default_factory=list,
        description=(
            "Compliance-relevant fields extracted by the upstream OCR/ML pipeline. "
            "All rule checks run against this list only. "
            "If a field_type is absent here but present in extracted_tags, the tag "
            "bridge will synthesize a field for it automatically."
        ),
    )
    extracted_tags: Optional[Dict[str, Any]] = Field(
        None,
        description=(
            "Structured key-value tags from the upstream ML service "
            "(MANUFACTURER_NAME, IMPORTER_NAME, NET_QUANTITY, MRP, etc.). "
            "The engine's tag bridge converts these into ExtractedField objects "
            "for any field_type not already covered by extracted_fields."
        ),
    )
    text_measurements: Optional[List[TextMeasurementEntry]] = Field(
        None,
        description=(
            "Per-text-block font height measurements from the OCR service. "
            "Used to attach font_height_mm to tag-synthesized fields when available."
        ),
    )

    @field_validator("package_shape")
    @classmethod
    def normalize_package_shape(cls, v: str) -> str:
        """Normalise casing — OCR service may send 'CYLINDRICAL', we need lowercase."""
        return v.lower()

    @field_validator("category")
    @classmethod
    def validate_category(cls, v: str) -> str:
        allowed = {
            # Standard product categories (also used for food-advisory logic)
            "food",
            "cosmetics",
            "electronics",
            "other",
            # Bulk-commodity categories with higher exemption thresholds (>50 kg/L)
            # These are distinct from 'other' so applicability.py can look up the
            # correct threshold from rules_config.json without guessing.
            "cement",
            "fertilizer",
            "agricultural_produce",
        }
        if v not in allowed:
            raise ValueError(f"category must be one of {sorted(allowed)}, got '{v}'")
        return v

    @field_validator("package_surface_type")
    @classmethod
    def validate_surface_type(cls, v: str) -> str:
        allowed = {"printed", "embossed_molded_blown"}
        if v not in allowed:
            raise ValueError(
                f"package_surface_type must be one of {sorted(allowed)}, got '{v}'"
            )
        return v

    @field_validator("sale_channel")
    @classmethod
    def validate_sale_channel(cls, v: str) -> str:
        allowed = {"retail", "ecommerce"}
        if v not in allowed:
            raise ValueError(
                f"sale_channel must be one of {sorted(allowed)}, got '{v}'"
            )
        return v


# ---------------------------------------------------------------------------
# Output models
# ---------------------------------------------------------------------------

class ViolationResult(BaseModel):
    """
    A single compliance violation discovered by one rule.
    Every violation is traceable to an exact legal clause.
    """

    rule_id: str = Field(..., description="Unique identifier of the rule that fired")
    clause: str = Field(
        ...,
        description="Exact legal citation, e.g. 'Rule 6(1)(a)' or 'Rule 7'",
    )
    severity: str = Field(
        ...,
        description="Impact level: 'major' (prosecution-level), 'minor', or 'advisory' (food-category downgrade)",
    )
    description: str = Field(
        ..., description="Human-readable explanation of what is wrong"
    )
    field_evidence: Optional[str] = Field(
        None,
        description="The raw field text that triggered this violation (for audit trail)",
    )
    field_type: Optional[str] = Field(
        None, description="Which field_type this violation relates to"
    )


class ApplicabilityNote(BaseModel):
    """Explains why a particular rule was or was not applied."""

    rule_id: str
    applied: bool
    reason: str
    severity_override: Optional[str] = None  # e.g. "advisory" for food category


class ValidationResponse(BaseModel):
    """
    Complete structured compliance report returned by POST /validate.
    Designed for legal traceability — every violation cites its exact clause.
    """

    product_id: str
    ruleset_version: str = Field(
        ..., description="Semantic version of the ruleset used, e.g. '2011-r6-v1.0'"
    )
    overall_status: str = Field(
        ...,
        description=(
            "Aggregated status: 'compliant' | 'non_compliant' | 'exempt' | "
            "'requires_manual_review' (violations=0 but some fields need human check)"
        ),
    )
    exempt_reason: Optional[str] = Field(
        None,
        description="Set when overall_status is 'exempt', e.g. 'exempt_bulk_package'",
    )
    violations: List[ViolationResult] = Field(
        default_factory=list,
        description="All violations found; empty list means fully compliant",
    )
    manual_review_required: List[str] = Field(
        default_factory=list,
        description=(
            "field_type values whose extraction confidence was below threshold — "
            "a human inspector must verify these fields before a final compliance decision"
        ),
    )
    applicability_notes: List[ApplicabilityNote] = Field(
        default_factory=list,
        description="Explains which rules were skipped/downgraded and why",
    )
    total_violations: int = Field(0, description="Count of violations found")
    major_violations: int = Field(0, description="Count of major-severity violations")
    minor_violations: int = Field(0, description="Count of minor-severity violations")
    advisory_violations: int = Field(
        0,
        description=(
            "Count of advisory-severity violations (food-category downgrades). "
            "Invariant: major + minor + advisory == total_violations always holds."
        ),
    )
    synthesized_fields_summary: List[Dict[str, Any]] = Field(
        default_factory=list,
        description=(
            "Fields the engine synthesized automatically from extracted_tags. "
            "Each entry records {field_type, text, source_tags} so callers can "
            "audit what the tag bridge contributed and verify it matched their data. "
            "Empty when no tag synthesis occurred."
        ),
    )

    def model_post_init(self, __context: Any) -> None:
        self.total_violations = len(self.violations)
        self.major_violations = sum(1 for v in self.violations if v.severity == "major")
        self.minor_violations = sum(1 for v in self.violations if v.severity == "minor")
        self.advisory_violations = sum(1 for v in self.violations if v.severity == "advisory")


class RuleInfo(BaseModel):
    """Metadata about a single registered rule — returned by GET /rules."""

    rule_id: str
    clause: str
    severity: str
    description: str
    rule_class: str


class RuleRegistryResponse(BaseModel):
    """Response for GET /rules — lists all active rules and version info."""

    ruleset_version: str
    ruleset_description: str
    rules: List[RuleInfo]
