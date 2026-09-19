"""
rules/base.py — Abstract base class for all rule implementations.

Every rule in the Legal Metrology Rule Engine inherits from this class.
The interface guarantees:
  - A unique rule_id
  - An exact legal clause citation
  - A severity level (major | minor)
  - A pure check() method that is independently testable with no shared mutable state
"""

from __future__ import annotations

from abc import ABC, abstractmethod
from typing import List, Optional, TYPE_CHECKING

if TYPE_CHECKING:
    from schemas import ExtractedField, ValidationRequest, ViolationResult


class Rule(ABC):
    """
    Abstract base for all Legal Metrology rule checks.

    Subclasses must set class-level attributes and implement check().
    The check() method must be a PURE FUNCTION — given the same inputs,
    it always returns the same output with no side effects.
    """

    #: Unique machine-readable identifier for this rule
    rule_id: str

    #: Exact legal citation from the Legal Metrology (Packaged Commodities) Rules, 2011
    clause: str

    #: Impact classification — 'major' means prosecution-level, 'minor' is advisory
    severity: str

    @abstractmethod
    def check(
        self,
        fields: List["ExtractedField"],
        context: "ValidationRequest",
    ) -> Optional["ViolationResult"]:
        """
        Run the compliance check against the extracted fields.

        Args:
            fields: High-confidence extracted fields from the product label.
                    Low-confidence fields are already filtered out by the engine.
            context: Full validation request containing product metadata.

        Returns:
            A ViolationResult if the rule is violated, None if compliant.
            Some rules may return multiple violations (e.g., font size checks
            per field) — those rules can be extended to return List[ViolationResult]
            but the default contract is Optional[ViolationResult].
        """
        ...

    def _get_fields_by_type(
        self, fields: List["ExtractedField"], field_type: str
    ) -> List["ExtractedField"]:
        """Helper: filter fields by field_type."""
        return [f for f in fields if f.field_type == field_type]

    def _first_field(
        self, fields: List["ExtractedField"], field_type: str
    ) -> Optional["ExtractedField"]:
        """Helper: return first field of given type, or None."""
        matches = self._get_fields_by_type(fields, field_type)
        return matches[0] if matches else None
