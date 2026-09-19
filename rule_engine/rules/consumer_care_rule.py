"""
rules/consumer_care_rule.py

Legal Basis: Rule 6(1)(f) of the Legal Metrology (Packaged Commodities) Rules, 2011,
as inserted by the 2017 Amendment.

Statutory text (paraphrased):
  Every package shall bear the name, address, and telephone number or e-mail address
  of the person or office who can be, or could be, contacted in case of consumer
  complaints.

Plain English:
  The label must display a contact point for consumer complaints — this is the
  "consumer care" or "consumer helpline" detail. It must contain either:
  - A valid 10-digit Indian mobile number, OR a valid STD landline number, OR
  - A valid e-mail address.
  The field must be present; a generic "For queries, call us" without an actual
  number or email is a violation.
"""

from __future__ import annotations

import json
import re
from pathlib import Path
from typing import List, Optional

from rules.base import Rule
from schemas import ExtractedField, ValidationRequest, ViolationResult

# Load contact patterns from central config
_CONFIG_PATH = Path(__file__).parent.parent / "rules_config.json"
with _CONFIG_PATH.open(encoding="utf-8") as _f:
    _CONFIG = json.load(_f)

_PHONE_PATTERN: re.Pattern = re.compile(
    _CONFIG["consumer_care_patterns"]["phone_regex"], re.IGNORECASE
)
_EMAIL_PATTERN: re.Pattern = re.compile(
    _CONFIG["consumer_care_patterns"]["email_regex"], re.IGNORECASE
)

# Toll-free / 1800 patterns also accepted
_TOLLFREE_PATTERN: re.Pattern = re.compile(r"1[- ]?800[- ]?\d{3,7}", re.IGNORECASE)


class ConsumerCareRule(Rule):
    """
    Rule 6(1)(f) — Consumer Care Contact Information.

    Checks that:
    1. A 'consumer_care' field is present on the label.
    2. The field contains a valid Indian phone number (10-digit mobile or STD landline
       or 1800 toll-free), OR a valid e-mail address pattern.

    Severity: minor (important for consumer protection but lower prosecution risk
    compared to absent MRP or quantity).
    """

    rule_id: str = "CONSUMER_CARE_001"
    clause: str = "Rule 6(1)(f)"
    severity: str = "minor"

    def check(
        self,
        fields: List[ExtractedField],
        context: ValidationRequest,
    ) -> Optional[ViolationResult]:
        """
        Validates consumer care contact details on the label.

        Returns ViolationResult for: missing field or field with no valid
        phone/email contact info. Returns None if compliant.
        """
        care_fields = self._get_fields_by_type(fields, "consumer_care")

        # --- Check 1: Field must be present ---
        if not care_fields:
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    "Consumer care contact information is missing from the label. "
                    "Rule 6(1)(f) (as amended in 2017) requires every package to display "
                    "a telephone number or e-mail address for consumer complaints."
                ),
                field_evidence=None,
                field_type="consumer_care",
            )

        primary_field = care_fields[0]
        text = primary_field.text.strip()

        # --- Check 2: Must contain a valid phone or email ---
        has_phone = bool(
            _PHONE_PATTERN.search(text) or _TOLLFREE_PATTERN.search(text)
        )
        has_email = bool(_EMAIL_PATTERN.search(text))

        if not has_phone and not has_email:
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    f"Consumer care field '{text}' does not contain a valid contact "
                    "detail. Rule 6(1)(f) requires a 10-digit Indian mobile number, "
                    "a valid STD landline, a 1800 toll-free number, or an e-mail address. "
                    "A generic statement without actionable contact information is insufficient."
                ),
                field_evidence=text,
                field_type="consumer_care",
            )

        return None  # Compliant
