"""
rules/generic_name_rule.py

Legal Basis: Rule 6(1)(a) read with Rule 2(l) of the Legal Metrology (Packaged
Commodities) Rules, 2011.

Statutory text (paraphrased):
  Rule 2(l): "generic name or common name" means the name used to describe the
  essential character of the commodity.
  Rule 6(1)(a): Every package shall bear a declaration of the generic or common name
  of the commodity contained in it.

Plain English:
  The label must tell the consumer WHAT the product fundamentally is — not just a
  brand name. "Sunrise" as the only name on a packet of biscuits is a violation;
  the word "Biscuits" (or equivalent) must also appear as the generic name.
  Brand names and trademarks do NOT satisfy this requirement.
"""

from __future__ import annotations

from typing import List, Optional

from rules.base import Rule
from schemas import ExtractedField, ValidationRequest, ViolationResult

_MIN_TEXT_LENGTH = 2  # Must be at least a meaningful word, not a single character


class GenericNameRule(Rule):
    """
    Rule 6(1)(a) & Rule 2(l) — Generic or Common Name of the Commodity.

    Checks that:
    1. A 'generic_name' field is present on the label.
    2. The field contains non-trivial text (at least 2 characters after stripping).

    A brand name alone does NOT satisfy Rule 6(1)(a). The generic/common name
    must appear as a separate declaration identifying the nature of the product
    (e.g., "Biscuits", "Refined Sunflower Oil", "USB-C Cable").

    Severity: major — consumers cannot make informed purchasing decisions without
    knowing what the product fundamentally is.
    """

    rule_id: str = "GENERIC_NAME_001"
    clause: str = "Rule 6(1)(a) & Rule 2(l)"
    severity: str = "major"

    def check(
        self,
        fields: List[ExtractedField],
        context: ValidationRequest,
    ) -> Optional[ViolationResult]:
        """
        Validates generic/common name declaration on the label.

        Returns ViolationResult for: missing field or field with trivially
        short text. Returns None if compliant.
        """
        generic_name_fields = self._get_fields_by_type(fields, "generic_name")

        # --- Check 1: Field must be present ---
        if not generic_name_fields:
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    "Generic or common name of the commodity is missing from the label. "
                    "Rule 6(1)(a) read with Rule 2(l) requires every package to declare "
                    "the generic or common name of the product (e.g., 'Biscuits', "
                    "'Refined Sunflower Oil', 'LED Bulb'). A brand or trade name alone "
                    "does not satisfy this requirement."
                ),
                field_evidence=None,
                field_type="generic_name",
            )

        primary_field = generic_name_fields[0]
        text = primary_field.text.strip()

        # --- Check 2: Text must be non-trivial ---
        if len(text) < _MIN_TEXT_LENGTH:
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    f"Generic name field contains insufficient text: '{text}'. "
                    "A meaningful generic or common name identifying the nature of "
                    "the commodity must be declared under Rule 6(1)(a) & Rule 2(l)."
                ),
                field_evidence=text,
                field_type="generic_name",
            )

        return None  # Compliant
