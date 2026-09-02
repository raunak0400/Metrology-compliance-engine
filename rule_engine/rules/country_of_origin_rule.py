"""
rules/country_of_origin_rule.py

Legal Basis: Rule 6(1)(b) proviso of the Legal Metrology (Packaged Commodities)
Rules, 2011, read with the Country of Origin (Compulsory Indication) Order, 1983.

Statutory text (paraphrased):
  Where any commodity is imported into India, every package of such commodity shall
  also bear a declaration as to the name of the country where the commodity is
  originally manufactured or produced.

Plain English:
  This rule ONLY applies to imported products. If is_imported is False, the rule
  engine (via applicability.py) will skip this check entirely.
  When the product IS imported, the label must display the country of origin
  (e.g., "Made in China", "Product of USA"). An absent or empty field is a violation.
"""

from __future__ import annotations

from typing import List, Optional

from rules.base import Rule
from schemas import ExtractedField, ValidationRequest, ViolationResult

_MIN_TEXT_LENGTH = 2  # At minimum a country code like "CN" must appear


class CountryOfOriginRule(Rule):
    """
    Rule 6(1)(b) proviso — Country of Origin for Imported Goods.

    Checks that:
    1. A 'country_of_origin' field is present (only runs when is_imported=True).
    2. The field contains non-trivial text identifying the source country.

    This rule is conditionally applied by the applicability engine:
    it is ONLY activated when ValidationRequest.is_imported == True.
    If is_imported is False, this rule should never be called (but includes
    a defensive check for belt-and-braces reliability).

    Severity: major (importing without country of origin marking violates
    both LM Rules and Customs requirements).
    """

    rule_id: str = "COUNTRY_ORIGIN_001"
    clause: str = "Rule 6(1)(b) proviso"
    severity: str = "major"

    def check(
        self,
        fields: List[ExtractedField],
        context: ValidationRequest,
    ) -> Optional[ViolationResult]:
        """
        Validates country of origin declaration for imported goods.

        Belt-and-braces guard: returns None immediately if is_imported is False.
        Otherwise checks field presence and non-trivial text content.
        """
        # Defensive guard — applicability should have excluded this for domestic products
        if not context.is_imported:
            return None

        origin_fields = self._get_fields_by_type(fields, "country_of_origin")

        # --- Check 1: Field must be present ---
        if not origin_fields:
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    "Country of origin declaration is missing from the label. "
                    "Rule 6(1)(b) proviso requires that every imported package bears "
                    "a clear declaration of the country where the commodity was "
                    "manufactured or produced (e.g., 'Made in China', 'Product of USA')."
                ),
                field_evidence=None,
                field_type="country_of_origin",
            )

        primary_field = origin_fields[0]
        text = primary_field.text.strip()

        # --- Check 2: Text must be non-trivial ---
        if len(text) < _MIN_TEXT_LENGTH:
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    f"Country of origin field contains insufficient text: '{text}'. "
                    "A recognisable country name or standard country code is required."
                ),
                field_evidence=text,
                field_type="country_of_origin",
            )

        return None  # Compliant
