"""
rules/placement_rule.py

Legal Basis: Rule 9 of the Legal Metrology (Packaged Commodities) Rules, 2011.

Statutory text (paraphrased):
  Rule 9: All declarations required to be made on a package under these rules shall
  appear on the principal display panel of the package. The principal display panel
  means the part of a package that is most likely to be displayed or presented to
  the consumer at the point of sale.

Plain English:
  Key mandatory declarations (MRP, net quantity, generic name, manufacturer) MUST
  appear on the main face of the package — the side that faces the consumer in a
  retail setting. Burying them on the bottom, back, or a side panel that faces
  the shelf is a violation.

  This rule is distinct from the font size rule — even a large-font field placed
  on the wrong panel is a separate, independent violation.
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import List, Optional

from rules.base import Rule
from schemas import ExtractedField, ValidationRequest, ViolationResult

# Load which field types must appear on the principal display panel
_CONFIG_PATH = Path(__file__).parent.parent / "rules_config.json"
with _CONFIG_PATH.open(encoding="utf-8") as _f:
    _CONFIG = json.load(_f)

_PLACEMENT_REQUIRED_FIELDS: List[str] = _CONFIG[
    "placement_mandatory_on_principal_panel"
]
_PRINCIPAL_PANEL = "principal_display_panel"


class PlacementRule(Rule):
    """
    Rule 9 — Principal Display Panel Placement.

    Checks that:
    1. Each field type that must appear on the principal display panel
       (as listed in rules_config.json) is actually located there.
    2. A field with panel_location == 'other' triggers a violation.

    Severity: minor (placement violations are enforced but carry lower
    prosecution risk than quantity/MRP violations).

    Like FontSizeRule, this rule may produce multiple violations (one per
    misplaced field). Use check_all() for full coverage.
    """

    rule_id: str = "PLACEMENT_001"
    clause: str = "Rule 9"
    severity: str = "minor"

    def check(
        self,
        fields: List[ExtractedField],
        context: ValidationRequest,
    ) -> Optional[ViolationResult]:
        """Returns the first placement violation found, or None if compliant."""
        violations = self.check_all(fields, context)
        return violations[0] if violations else None

    def check_all(
        self,
        fields: List[ExtractedField],
        context: ValidationRequest,
    ) -> List[ViolationResult]:
        """
        Validates panel placement for all mandatory fields.

        Returns a list of ViolationResult, one per field found on the wrong panel.
        An empty list means all mandatory fields are correctly placed.
        """
        violations: List[ViolationResult] = []

        for field in fields:
            if field.field_type not in _PLACEMENT_REQUIRED_FIELDS:
                continue  # This field type is not required on principal panel

            if field.panel_location != _PRINCIPAL_PANEL:
                violations.append(
                    ViolationResult(
                        rule_id=self.rule_id,
                        clause=self.clause,
                        severity=self.severity,
                        description=(
                            f"Field '{field.field_type}' is declared on panel "
                            f"'{field.panel_location}' but must appear on the "
                            f"'{_PRINCIPAL_PANEL}' under Rule 9. "
                            f"Field text: '{field.text[:60]}'"
                        ),
                        field_evidence=(
                            f"panel_location='{field.panel_location}', "
                            f"text='{field.text[:40]}'"
                        ),
                        field_type=field.field_type,
                    )
                )

        return violations
