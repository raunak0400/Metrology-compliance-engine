"""
tests/test_manufacturer_rule.py

Unit tests for ManufacturerRule — Rule 6(1)(a).
Tests cover: field presence, text length, imported product importer check.
"""

from __future__ import annotations

import sys
import os

# Ensure rule_engine root is on the path when running from tests/
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import pytest
from schemas import ExtractedField, ValidationRequest, BoundingBox
from rules.manufacturer_rule import ManufacturerRule
from tests.test_fixtures import (
    FIXTURE_COMPLIANT_RETAIL,
    FIXTURE_MISSING_MRP,
    FIXTURE_IMPORTED_MISSING_COUNTRY,
    FIXTURE_IMPORTED_WITH_COUNTRY,
)


def _make_field(field_type: str, text: str, confidence: float = 0.95) -> ExtractedField:
    return ExtractedField(
        field_type=field_type,
        text=text,
        confidence=confidence,
        bbox=BoundingBox(x=0, y=0, w=100, h=20),
        font_height_mm=1.5,
        panel_location="principal_display_panel",
    )


def _make_request(fixture: dict) -> ValidationRequest:
    return ValidationRequest(**fixture)


RULE = ManufacturerRule()


class TestManufacturerRulePresence:
    """Tests for manufacturer field presence."""

    def test_compliant_domestic(self):
        """Fully compliant domestic product — no violation."""
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        fields = [f for f in req.extracted_fields if f.confidence >= 0.75]
        result = RULE.check(fields, req)
        assert result is None, "Expected no violation for compliant domestic product"

    def test_missing_manufacturer_field(self):
        """No manufacturer field at all — should violate Rule 6(1)(a)."""
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        # Remove manufacturer fields
        fields = [f for f in req.extracted_fields if f.field_type != "manufacturer"]
        result = RULE.check(fields, req)
        assert result is not None
        assert result.rule_id == "MANUFACTURER_001"
        assert result.clause == "Rule 6(1)(a)"
        assert result.severity == "major"
        assert result.field_type == "manufacturer"

    def test_empty_manufacturer_text(self):
        """Manufacturer field present but essentially empty — violation."""
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        fields = [_make_field("manufacturer", "   ")]  # whitespace only
        result = RULE.check(fields, req)
        assert result is not None
        assert "insufficient" in result.description.lower() or "missing" in result.description.lower()

    def test_too_short_manufacturer_text(self):
        """Text < 5 chars — violation."""
        fields = [_make_field("manufacturer", "AB")]
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        result = RULE.check(fields, req)
        assert result is not None
        assert result.rule_id == "MANUFACTURER_001"

    def test_adequate_manufacturer_text(self):
        """Text >= 5 chars — compliant."""
        fields = [_make_field("manufacturer", "Acme Industries Pvt. Ltd., Mumbai")]
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        result = RULE.check(fields, req)
        assert result is None


class TestManufacturerRuleImported:
    """Tests for imported product importer reference requirements."""

    def test_imported_missing_importer_keyword(self):
        """Imported product without importer keyword — violation."""
        req = _make_request(FIXTURE_IMPORTED_MISSING_COUNTRY)
        # Give it a manufacturer field WITHOUT any importer keyword
        fields = [_make_field("manufacturer", "Overseas Manufacturer Co., Shanghai")]
        result = RULE.check(fields, req)
        assert result is not None
        assert "import" in result.description.lower()

    def test_imported_with_importer_keyword(self):
        """Imported product with 'Importer' keyword — compliant."""
        req = _make_request(FIXTURE_IMPORTED_WITH_COUNTRY)
        fields = [
            _make_field(
                "manufacturer",
                "TechImports India Pvt Ltd (Importer), 101 Import Hub, Mumbai",
            )
        ]
        result = RULE.check(fields, req)
        assert result is None

    def test_imported_with_two_fields(self):
        """Imported product with two manufacturer fields (manufacturer + importer) — compliant."""
        req = _make_request(FIXTURE_IMPORTED_WITH_COUNTRY)
        fields = [
            _make_field("manufacturer", "Factory Co., Shenzhen, China"),
            _make_field("manufacturer", "LocalDistrib Pvt Ltd, Mumbai"),
        ]
        result = RULE.check(fields, req)
        assert result is None

    def test_domestic_product_no_importer_check(self):
        """Domestic product — importer check not triggered even without keyword."""
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        fields = [_make_field("manufacturer", "Purely Domestic Maker, Pune")]
        result = RULE.check(fields, req)
        assert result is None


class TestManufacturerRuleMetadata:
    """Tests for rule metadata correctness."""

    def test_rule_id(self):
        assert RULE.rule_id == "MANUFACTURER_001"

    def test_clause(self):
        assert RULE.clause == "Rule 6(1)(a)"

    def test_severity(self):
        assert RULE.severity == "major"
