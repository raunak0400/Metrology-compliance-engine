"""
rules/mfg_date_rule.py

Legal Basis: Rule 6(1)(d) of the Legal Metrology (Packaged Commodities) Rules, 2011.

Statutory text (paraphrased):
  Every package shall bear a declaration of the month and year in which the commodity
  is manufactured or packed or imported.

Note on Rule 6(10) / 2017 Amendment:
  Rule 6(10) requires that e-commerce ONLINE LISTINGS clearly display all mandatory
  declarations. It ADDS an obligation for the platform — it does NOT remove the
  mfg_date requirement from the physical label. The physical label of a product
  sold e-commerce must still carry a month-year manufacture/packing declaration.

Plain English:
  The label must show at minimum the month and year the product was made or packed.
  A full date is fine; month-year is the minimum. Year-only is a violation.
  Future dates are flagged as suspicious. The rule engine accepts common date formats
  such as "Jan 2024", "01/2024", "January, 2024", "03-2024".
"""

from __future__ import annotations

import json
import re
from datetime import datetime
from pathlib import Path
from typing import List, Optional

from rules.base import Rule
from schemas import ExtractedField, ValidationRequest, ViolationResult

# Load date format patterns from config (single source of truth)
_CONFIG_PATH = Path(__file__).parent.parent / "rules_config.json"
with _CONFIG_PATH.open(encoding="utf-8") as _f:
    _CONFIG = json.load(_f)

_DATE_PATTERNS: List[re.Pattern] = [
    re.compile(p, re.IGNORECASE) for p in _CONFIG["mfg_date_patterns"]
]

# Year-only pattern — accepted in some contexts but insufficient under Rule 6(1)(d)
_YEAR_ONLY_PATTERN = re.compile(r"^\s*\d{4}\s*$")

_CURRENT_YEAR = datetime.now().year
_CURRENT_MONTH = datetime.now().month


class MfgDateRule(Rule):
    """
    Rule 6(1)(d) — Month and Year of Manufacture / Packing / Import.

    Checks that:
    1. A 'mfg_date' field is present on the label.
    2. The field contains a month+year date pattern (not year-only).
    3. The date is not in the future (a suspicious indicator).

    Note: This rule is skipped entirely for e-commerce sale channels —
    the applicability engine removes it before execution. The defensive
    check below is belt-and-braces only.

    Severity: major (consumers have a legal right to know production date).
    """

    rule_id: str = "MFG_DATE_001"
    clause: str = "Rule 6(1)(d)"
    severity: str = "major"

    def check(
        self,
        fields: List[ExtractedField],
        context: ValidationRequest,
    ) -> Optional[ViolationResult]:
        """
        Validates manufacture/packing date field presence and format.

        Returns ViolationResult for: missing field, year-only format,
        or future date. Returns None if compliant.

        Note: sale_channel does NOT affect this rule. Rule 6(10) adds an
        online display obligation for e-commerce platforms; it does not exempt
        the physical label from the mfg_date declaration.
        """
        date_fields = self._get_fields_by_type(fields, "mfg_date")

        # --- Check 1: Field must be present ---
        if not date_fields:
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    "Month and year of manufacture/packing is missing from the label. "
                    "Rule 6(1)(d) requires every package to declare at minimum the month "
                    "and year in which the commodity was manufactured, packed, or imported."
                ),
                field_evidence=None,
                field_type="mfg_date",
            )

        primary_field = date_fields[0]
        text = primary_field.text.strip()

        # --- Check 2: Year-only is insufficient ---
        if _YEAR_ONLY_PATTERN.match(text):
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    f"Manufacture date '{text}' provides only a year. "
                    "Rule 6(1)(d) requires at minimum the month AND year to be declared."
                ),
                field_evidence=text,
                field_type="mfg_date",
            )

        # --- Check 3: Must match a known month+year pattern ---
        matched_pattern = any(pat.search(text) for pat in _DATE_PATTERNS)
        if not matched_pattern:
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    f"Manufacture date field '{text}' does not match any recognised "
                    "month-year format (e.g., 'Jan 2024', '01/2024', 'January, 2024', "
                    "'03-2024'). Rule 6(1)(d) requires a clear month and year declaration."
                ),
                field_evidence=text,
                field_type="mfg_date",
            )

        # --- Check 4: Date must not be in the future ---
        # Extract 4-digit year from text for a basic future-date sanity check
        year_match = re.search(r"\b(20\d{2}|19\d{2})\b", text)
        if year_match:
            declared_year = int(year_match.group(1))
            if declared_year > _CURRENT_YEAR:
                return ViolationResult(
                    rule_id=self.rule_id,
                    clause=self.clause,
                    severity="minor",  # Suspicious but not always malicious
                    description=(
                        f"Manufacture date '{text}' appears to be in the future "
                        f"(year {declared_year} > current year {_CURRENT_YEAR}). "
                        "This may indicate a labelling error or pre-printed stock."
                    ),
                    field_evidence=text,
                    field_type="mfg_date",
                )

        return None  # Compliant
