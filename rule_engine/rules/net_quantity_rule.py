"""
rules/net_quantity_rule.py

Legal Basis: Rule 6(1)(c) and Rule 11 of the Legal Metrology (Packaged Commodities)
Rules, 2011.

Statutory text (paraphrased):
  Rule 6(1)(c): Every package shall bear a declaration of the net quantity in terms of
  standard unit of weights or measures.
  Rule 11: Net quantity shall be expressed in metric system — grams/kilograms for weight,
  millilitres/litres for volume, metres for length, etc.

Plain English:
  The label must state how much product is in the package, using a recognised metric
  unit. "500 grams", "1 litre", "12 pcs" are all valid. A unit like "handful" or an
  unrecognised abbreviation is a violation. An absent quantity field is also a violation.
"""

from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Dict, List, Optional, Set

from rules.base import Rule
from schemas import ExtractedField, ValidationRequest, ViolationResult

# Load allowed units from central config
_CONFIG_PATH = Path(__file__).parent.parent / "rules_config.json"
with _CONFIG_PATH.open(encoding="utf-8") as _f:
    _CONFIG = json.load(_f)

# Flatten all allowed unit variants into a single set (lower-cased)
_ALLOWED_UNITS: Set[str] = set()
for _category_units in _CONFIG["net_quantity_allowed_units"].values():
    for _unit in _category_units:
        _ALLOWED_UNITS.add(_unit.lower())


def _extract_unit_from_text(text: str) -> Optional[str]:
    """
    Parse out the unit token from a net-quantity string.

    Handles:
      - Simple:      '500 g', '1.5 kg', '200 ml', '12 pcs'
      - Multi-pack:  '2 x 325 ml', '6 x 100 g', '2 x 325 ml (Total: 650 ml)'
      - Bonus-pack:  '300 ml + 30 ml FREE', '500 g + 50 g free'
      - Compound:    '2 sq.m', '5 sq ft'

    Strategy (ordered by priority):
      0. Multi-pack "N x qty unit" — extracts unit from the PER-UNIT part,
         not the multiplier. Runs first to prevent 'x' being tokenised as unit.
      1. Compound area units (sq.m, sq ft) — longest match, must beat path 2.
      2. Bonus-pack "qty unit + qty unit" — extracts unit from the base qty.
      3. Strict single-token unit anchored to whitespace/end-of-string.

    Returns the raw unit substring (lowercased), or None if no pattern matches.
    """
    # --- Path 0: Multi-pack notation "N x qty unit" ---
    # Examples: "2 x 325 ml", "6 x 100 g", "2 x 325 ml (Total: 650 ml)"
    # The multiplier (N) and separator (x/X/×) are consumed; unit is taken
    # from the per-unit quantity that follows.
    multipack_pattern = re.compile(
        r"^\s*\d+\s*[xX×]\s*[\d,]+(?:\.\d+)?\s*([a-zA-Z]+\.?)(?:\s|$|[,.(])",
        re.IGNORECASE,
    )
    multipack_match = multipack_pattern.match(text.strip())
    if multipack_match:
        return multipack_match.group(1).rstrip(".").lower()

    # --- Path 1: Compound area units (must match before single-token path) ---
    compound_pattern = re.compile(
        r"[\d,]+(?:\.\d+)?\s*(sq\.?\s?m|sq\.?\s?ft)", re.IGNORECASE
    )
    compound_match = compound_pattern.search(text)
    if compound_match:
        return compound_match.group(1).replace(" ", "").rstrip(".").lower()

    # --- Path 2: Bonus-pack "qty unit + qty unit FREE" ---
    # Examples: "300 ml + 30 ml FREE", "500 g + 50 g free"
    # Extracts unit from the FIRST (base) quantity only.
    bonus_pattern = re.compile(
        r"([\d,]+(?:\.\d+)?)\s*([a-zA-Z]+)\s*\+\s*[\d,]+(?:\.\d+)?\s*[a-zA-Z]*\s*(?:free|FREE)?",
        re.IGNORECASE,
    )
    bonus_match = bonus_pattern.match(text.strip())
    if bonus_match:
        unit = bonus_match.group(2).rstrip(".").lower()
        return unit

    # --- Path 3: Strict single-token unit anchored to word boundary ---
    # The (?:\s|$|[,.]) anchor prevents trailing words being absorbed into
    # the unit match (e.g. "500 g Pack" must not yield "gpack").
    pattern = re.compile(
        r"[\d,]+(?:\.\d+)?\s*([a-zA-Z]+\.?)(?:\s|$|[,.(])"
    )
    match = pattern.search(text)
    if not match:
        return None
    return match.group(1).rstrip(".").lower()



class NetQuantityRule(Rule):
    """
    Rule 6(1)(c) & Rule 11 — Net Quantity Declaration.

    Checks that:
    1. A 'net_quantity' field is present on the label.
    2. The field contains a numeric value followed by a valid metric unit.
    3. The unit belongs to the allowed list defined in rules_config.json.

    Severity: major (incorrect/missing quantity is directly deceptive to consumers).
    """

    rule_id: str = "NET_QUANTITY_001"
    clause: str = "Rule 6(1)(c) & Rule 11"
    severity: str = "major"

    def check(
        self,
        fields: List[ExtractedField],
        context: ValidationRequest,
    ) -> Optional[ViolationResult]:
        """
        Validates net quantity field presence and unit compliance.

        Returns a ViolationResult for: missing field, non-numeric content,
        or unrecognised unit. Returns None if compliant.
        """
        quantity_fields = self._get_fields_by_type(fields, "net_quantity")

        # --- Check 1: Field presence ---
        if not quantity_fields:
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    "Net quantity declaration is missing from the label. "
                    "Rule 6(1)(c) requires every package to declare the net quantity "
                    "in standard units (weight, volume, length, or count)."
                ),
                field_evidence=None,
                field_type="net_quantity",
            )

        primary_field = quantity_fields[0]
        text = primary_field.text.strip()

        # --- Check 2: Must contain a numeric value ---
        if not re.search(r"[\d]", text):
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    f"Net quantity field contains no numeric value: '{text}'. "
                    "A quantity declaration must include a number and a standard unit."
                ),
                field_evidence=text,
                field_type="net_quantity",
            )

        # --- Check 3: Unit must be in the allowed list ---
        detected_unit = _extract_unit_from_text(text)
        if detected_unit is None:
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    f"Net quantity field '{text}' does not contain a recognisable unit. "
                    "Rule 11 requires quantity in metric system: g/kg for weight, "
                    "ml/l for volume, m/cm/mm for length, or a countable number (pcs/nos)."
                ),
                field_evidence=text,
                field_type="net_quantity",
            )

        if detected_unit not in _ALLOWED_UNITS:
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    f"Net quantity unit '{detected_unit}' (from '{text}') is not a "
                    f"recognised standard metric unit under Rule 11. Allowed units include: "
                    f"g, kg, ml, l, m, cm, mm, pcs, nos, and equivalents."
                ),
                field_evidence=text,
                field_type="net_quantity",
            )

        return None  # Compliant
