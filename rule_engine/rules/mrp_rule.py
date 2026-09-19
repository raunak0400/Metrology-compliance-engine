"""
rules/mrp_rule.py

Legal Basis: Rule 6(1)(e) of the Legal Metrology (Packaged Commodities) Rules, 2011,
as amended by the Legal Metrology (Packaged Commodities) Amendment Rules, 2017.

Statutory text (paraphrased):
  Every package shall bear a declaration of the Maximum Retail Price (inclusive of all
  taxes) at which the commodity in the package shall be sold to the ultimate consumer.
  The price shall be preceded by the words "Maximum Retail Price" or "MRP" and followed
  by the phrase "inclusive of all taxes" (or an accepted abbreviation thereof).

Plain English:
  The label must show the MRP and explicitly state that it includes all taxes. Just
  printing a price without the "incl. of all taxes" language is a violation. Similarly,
  having the phrase without a price is a violation. Both elements are mandatory.
"""

from __future__ import annotations

import json
import re
from pathlib import Path
from typing import List, Optional

from rules.base import Rule
from schemas import ExtractedField, ValidationRequest, ViolationResult

# Load required phrases from central config (single source of truth for legal language)
_CONFIG_PATH = Path(__file__).parent.parent / "rules_config.json"
with _CONFIG_PATH.open(encoding="utf-8") as _f:
    _CONFIG = json.load(_f)

_REQUIRED_PHRASES: List[str] = _CONFIG["mrp_required_phrases"]

# Build a compiled OR-regex from all accepted phrases (case-insensitive)
_PHRASE_PATTERN: re.Pattern = re.compile(
    "|".join(re.escape(phrase) for phrase in _REQUIRED_PHRASES),
    re.IGNORECASE,
)

# Numeric price pattern: optional Rs/INR symbol, digits, optional decimal
_PRICE_PATTERN: re.Pattern = re.compile(
    r"""
    (?:rs\.?\s*|₹\s*|inr\s*)?   # optional currency symbol/prefix
    (?:\d{1,6}(?:[,\s]\d{3})*    # integer part, optional thousands separators
    (?:\.\d{1,2})?)              # optional paise/cents
    """,
    re.VERBOSE | re.IGNORECASE,
)


class MRPRule(Rule):
    """
    Rule 6(1)(e) — Maximum Retail Price (inclusive of all taxes).

    Checks that:
    1. An 'mrp' field is present on the label.
    2. The field contains a numeric price value.
    3. The field contains an accepted 'inclusive of all taxes' phrase.
       Accepted variants are loaded from rules_config.json — editing that file
       changes behaviour without touching rule code.

    Severity: major (MRP violations are among the most prosecuted under LM Act).
    """

    rule_id: str = "MRP_001"
    clause: str = "Rule 6(1)(e)"
    severity: str = "major"

    def check(
        self,
        fields: List[ExtractedField],
        context: ValidationRequest,
    ) -> Optional[ViolationResult]:
        """
        Validates MRP field presence, numeric price, and mandatory tax-inclusive phrase.

        Returns:
            ViolationResult for missing field, missing price, or missing tax phrase.
            None if fully compliant.
        """
        mrp_fields = self._get_fields_by_type(fields, "mrp")

        # --- Check 1: MRP field must be present ---
        if not mrp_fields:
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    "Maximum Retail Price (MRP) declaration is missing from the label. "
                    "Rule 6(1)(e) requires every retail package to display the MRP "
                    "inclusive of all taxes."
                ),
                field_evidence=None,
                field_type="mrp",
            )

        primary_field = mrp_fields[0]
        text = primary_field.text.strip()

        # --- Check 2: Must contain a numeric price ---
        price_match = _PRICE_PATTERN.search(text)
        # price_match must actually contain digits, not just match empty string
        if not price_match or not re.search(r"\d", price_match.group()):
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    f"MRP field '{text}' does not contain a numeric price. "
                    "A valid MRP declaration must include the actual price value."
                ),
                field_evidence=text,
                field_type="mrp",
            )

        # --- Check 3: Must contain the tax-inclusive phrase ---
        if not _PHRASE_PATTERN.search(text):
            accepted = ", ".join(f'"{p}"' for p in _REQUIRED_PHRASES[:4])
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    f"MRP field '{text}' is missing the mandatory tax-inclusive phrase. "
                    f"Rule 6(1)(e) requires one of: {accepted} (or equivalent) to appear "
                    f"alongside the price. This language is mandatory and not optional."
                ),
                field_evidence=text,
                field_type="mrp",
            )

        return None  # Compliant
