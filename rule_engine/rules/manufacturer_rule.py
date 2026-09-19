"""
rules/manufacturer_rule.py

Legal Basis: Rule 6(1)(a) of the Legal Metrology (Packaged Commodities) Rules, 2011.

Statutory text (paraphrased):
  Every package shall bear a declaration of the name and address of the manufacturer,
  or where the manufacturer is not the packer, the name and address of the packer.
  For imported goods, the name and address of the importer must also appear.

Plain English:
  The label must show who made (or packed) the product, with a real address.
  For imported products, the Indian importer's name and address is ALSO required —
  this is a separate declaration from the manufacturer field and is checked by
  looking for a distinct field_type == "importer" entry in extracted_fields,
  OR for importer-keyword text within any manufacturer-type field.
  An empty, whitespace-only, or absent field is a violation.
"""

from __future__ import annotations

from typing import List, Optional

from rules.base import Rule
from schemas import ExtractedField, ValidationRequest, ViolationResult


class ManufacturerRule(Rule):
    """
    Rule 6(1)(a) — Manufacturer / Packer / Importer declaration.

    Checks that:
    1. A 'manufacturer' field is present and contains non-trivial text.
    2. For imported products: a distinct 'importer' field_type is present with
       non-trivial text, OR any of the manufacturer-type fields explicitly names
       an importer/distributor. Checks the 'importer' field_type FIRST.

    Severity: major (missing manufacturer/importer info is a prosecutable offence).
    """

    rule_id: str = "MANUFACTURER_001"
    clause: str = "Rule 6(1)(a)"
    severity: str = "major"

    # Minimum character length for a name/address to be considered non-trivial
    _MIN_TEXT_LENGTH: int = 5

    # Keywords inside manufacturer text that signal an importer reference
    _IMPORTER_KEYWORDS = frozenset({
        "import", "importer", "imported by", "marketed by", "distributor",
        "sole distributor", "indian agent",
    })

    def check(
        self,
        fields: List[ExtractedField],
        context: ValidationRequest,
    ) -> Optional[ViolationResult]:
        """
        Validates that manufacturer/packer name and address are declared.

        For imported products, also verifies that importer identity is resolvable
        via a distinct field_type == 'importer' entry OR importer keywords in any
        manufacturer-type field. Returns None if compliant.
        """
        manufacturer_fields = self._get_fields_by_type(fields, "manufacturer")

        # --- Check 1: Manufacturer field must be present ---
        if not manufacturer_fields:
            description = (
                "Manufacturer/packer name and address is missing from the label. "
                "Every package must declare the name and complete address of the "
                "manufacturer or packer under Rule 6(1)(a)."
            )
            if context.is_imported:
                description += (
                    " Additionally, for imported goods, the Indian importer's "
                    "name and address must also be declared."
                )
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=description,
                field_evidence=None,
                field_type="manufacturer",
            )

        # --- Check 2: Manufacturer field text must be non-trivial ---
        primary_field = manufacturer_fields[0]
        text_stripped = primary_field.text.strip()

        if len(text_stripped) < self._MIN_TEXT_LENGTH:
            return ViolationResult(
                rule_id=self.rule_id,
                clause=self.clause,
                severity=self.severity,
                description=(
                    f"Manufacturer/packer field is present but contains insufficient "
                    f"text ('{text_stripped}'). A full name and address is required "
                    f"under Rule 6(1)(a)."
                ),
                field_evidence=primary_field.text,
                field_type="manufacturer",
            )

        # --- Check 3: Imported products — importer declaration must exist ---
        # Priority order:
        #   a) Look for a distinct field_type == "importer" entry (most explicit)
        #   b) Fall back to keyword-scan across all manufacturer-type fields
        #   c) Accept 2+ manufacturer entries (some labels split mfr/importer)
        if context.is_imported:
            # (a) Explicit "importer" field_type — checked FIRST
            importer_fields = self._get_fields_by_type(fields, "importer")
            if importer_fields:
                importer_text = importer_fields[0].text.strip()
                if len(importer_text) >= self._MIN_TEXT_LENGTH:
                    return None  # Compliant — explicit importer field present

            # (b) Keyword scan across all manufacturer-type fields
            combined_mfr_text = " ".join(
                f.text.lower() for f in manufacturer_fields
            )
            has_importer_keyword = any(
                kw in combined_mfr_text for kw in self._IMPORTER_KEYWORDS
            )

            # (c) Multiple manufacturer entries — second likely represents importer
            has_multiple_entries = len(manufacturer_fields) >= 2

            if not has_importer_keyword and not has_multiple_entries:
                return ViolationResult(
                    rule_id=self.rule_id,
                    clause=self.clause,
                    severity=self.severity,
                    description=(
                        "Product is marked as imported but the label does not contain "
                        "the importer's name and address. Rule 6(1)(a) proviso requires "
                        "that imported packages declare the Indian importer's full details. "
                        "Provide a distinct field_type 'importer' entry, or include "
                        "importer keywords in the manufacturer field."
                    ),
                    field_evidence=primary_field.text,
                    field_type="importer",
                )

        return None  # Compliant
