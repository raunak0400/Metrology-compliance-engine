"""
rules/font_size_rule.py

Legal Basis: Rule 7 of the Legal Metrology (Packaged Commodities) Rules, 2011,
as amended by the Legal Metrology (Packaged Commodities) Amendment Rules, 2017.

Statutory scope (Rule 7 — "Numeral size"):
  Rule 7 prescribes minimum HEIGHT REQUIREMENTS FOR NUMERALS used in the
  NET QUANTITY declaration. It does NOT set a numeric mm threshold for all
  mandatory declarations — those other fields fall under Rule 6's general
  "legible and conspicuous" standard, which has no attached mm figure.

  Rule 7(2): The height of every numeral in the NET QUANTITY declaration
  shall be not less than:
    (a) 1 mm for packages with a printed label
    (b) 2 mm for packages with embossed, moulded, or blown-in letters/numerals

Rule 6 (other fields):
  Manufacturer, MRP, mfg_date, consumer_care, generic_name fields must be
  "legible and clearly visible" — a qualitative standard with no statutory
  mm figure. These fields are NOT checked here.

Plain English:
  Only the net_quantity numeral height is subject to a measurable mm threshold
  under Rule 7. This rule checks only that field.
  Thresholds are loaded from rules_config.json — not hardcoded here.
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import List, Optional

from rules.base import Rule
from schemas import ExtractedField, ValidationRequest, ViolationResult

# Load font thresholds from central config — single source of truth
_CONFIG_PATH = Path(__file__).parent.parent / "rules_config.json"
with _CONFIG_PATH.open(encoding="utf-8") as _f:
    _CONFIG = json.load(_f)

_FONT_THRESHOLDS: dict = _CONFIG["font_size_thresholds_mm"]

# Rule 7 applies its numeric mm threshold to the NET QUANTITY numeral only.
# Other mandatory fields are assessed against Rule 6's qualitative "legible"
# standard — not a measurable mm threshold.
_RULE7_FIELD_TYPES = frozenset({"net_quantity"})


class FontSizeRule(Rule):
    """
    Rule 7 — Minimum numeral height for the net quantity declaration.

    Statutory scope: Rule 7 prescribes minimum numeral heights ONLY for the
    net quantity field. Manufacturer, MRP, and other Rule 6 declarations must
    be "legible" — a qualitative standard with no attached mm figure.

    Checks that:
    1. The 'net_quantity' field's font_height_mm meets the minimum for the
       package's surface type.
    2. Thresholds (from rules_config.json):
       - printed:               >= 1.0 mm
       - embossed_molded_blown: >= 2.0 mm
    3. Fields without a font_height_mm value are skipped (cannot assess).

    Severity: minor.

    Note: check_all() is provided for API consistency with the engine's
    MULTI_VIOLATION_RULES handling, but currently returns at most one result
    (one net_quantity field, one possible violation).
    """

    rule_id: str = "FONT_SIZE_001"
    clause: str = "Rule 7"
    severity: str = "minor"

    def _get_threshold_mm(self, context: ValidationRequest) -> float:
        """Return the applicable minimum numeral height in mm for this package type."""
        surface = context.package_surface_type
        return _FONT_THRESHOLDS.get(surface, _FONT_THRESHOLDS["printed"])

    def check(
        self,
        fields: List[ExtractedField],
        context: ValidationRequest,
    ) -> Optional[ViolationResult]:
        """
        Returns a font-size violation if the net_quantity numeral is too small.
        Returns None if compliant or if font_height_mm was not measured.
        """
        violations = self.check_all(fields, context)
        return violations[0] if violations else None

    def check_all(
        self,
        fields: List[ExtractedField],
        context: ValidationRequest,
    ) -> List[ViolationResult]:
        """
        Validates numeral height compliance for Rule 7-scoped fields only
        (currently: net_quantity).

        Returns a list of ViolationResult; empty list means compliant or
        font height was not measurable.
        """
        threshold_mm = self._get_threshold_mm(context)
        violations: List[ViolationResult] = []

        for field in fields:
            # Rule 7 only applies its mm threshold to net_quantity numerals.
            # All other mandatory fields are checked for legibility qualitatively
            # under Rule 6 — no numeric mm threshold applies to them.
            if field.field_type not in _RULE7_FIELD_TYPES:
                continue

            # Skip if font size was not measured by the upstream pipeline
            if field.font_height_mm is None:
                continue

            if field.font_height_mm < threshold_mm:
                violations.append(
                    ViolationResult(
                        rule_id=self.rule_id,
                        clause=self.clause,
                        severity=self.severity,
                        description=(
                            f"Net quantity numeral height is {field.font_height_mm:.2f} mm, "
                            f"below the Rule 7 minimum of {threshold_mm:.1f} mm for "
                            f"'{context.package_surface_type}' packages. "
                            f"Rule 7 mandates that net quantity numerals are clearly "
                            f"legible at this minimum height. "
                            f"Field text: '{field.text[:60]}'"
                        ),
                        field_evidence=(
                            f"font_height_mm={field.font_height_mm:.2f}, "
                            f"threshold={threshold_mm:.1f}, text='{field.text[:40]}'"
                        ),
                        field_type=field.field_type,
                    )
                )

        return violations
