"""
tests/test_mrp_rule.py

Unit tests for MRPRule — Rule 6(1)(e).
Tests cover: field presence, numeric price, tax-inclusive phrase variants.
"""

from __future__ import annotations

import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import pytest
from schemas import ExtractedField, ValidationRequest, BoundingBox
from rules.mrp_rule import MRPRule
from tests.test_fixtures import (
    FIXTURE_COMPLIANT_RETAIL,
    FIXTURE_MISSING_MRP,
    FIXTURE_MRP_MISSING_TAX_PHRASE,
)


def _make_field(text: str, confidence: float = 0.95) -> ExtractedField:
    return ExtractedField(
        field_type="mrp",
        text=text,
        confidence=confidence,
        bbox=BoundingBox(x=0, y=0, w=200, h=20),
        font_height_mm=1.5,
        panel_location="principal_display_panel",
    )


def _make_request(fixture: dict) -> ValidationRequest:
    return ValidationRequest(**fixture)


RULE = MRPRule()


class TestMRPRulePresence:
    """Tests for MRP field presence."""

    def test_compliant_mrp(self):
        """Standard compliant MRP — no violation."""
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        fields = [f for f in req.extracted_fields if f.confidence >= 0.75]
        result = RULE.check(fields, req)
        assert result is None

    def test_missing_mrp_field(self):
        """No MRP field — should violate Rule 6(1)(e)."""
        req = _make_request(FIXTURE_MISSING_MRP)
        fields = [f for f in req.extracted_fields if f.confidence >= 0.75]
        result = RULE.check(fields, req)
        assert result is not None
        assert result.rule_id == "MRP_001"
        assert result.clause == "Rule 6(1)(e)"
        assert result.severity == "major"
        assert result.field_type == "mrp"
        assert result.field_evidence is None  # field was absent


class TestMRPRuleTaxPhrase:
    """Tests for the mandatory 'inclusive of all taxes' phrase."""

    def test_mrp_missing_tax_phrase(self):
        """MRP without tax phrase — violation."""
        req = _make_request(FIXTURE_MRP_MISSING_TAX_PHRASE)
        fields = [f for f in req.extracted_fields if f.confidence >= 0.75]
        result = RULE.check(fields, req)
        assert result is not None
        assert "tax" in result.description.lower()
        assert result.field_evidence is not None

    @pytest.mark.parametrize("phrase,price,should_pass", [
        ("incl. of all taxes", "Rs. 50/-", True),
        ("inclusive of all taxes", "Rs. 100/-", True),
        ("incl of all taxes", "₹ 75/-", True),
        ("Incl. Of All Taxes", "Rs. 200/-", True),   # case-insensitive
        ("INCLUDING ALL TAXES", "Rs. 30/-", True),
        ("mrp (incl. of all taxes)", "Rs. 299/-", True),
        ("", "Rs. 50/-", False),                      # no phrase
        ("plus taxes", "Rs. 50/-", False),             # wrong phrase
        ("before taxes", "Rs. 50/-", False),           # wrong phrase
    ])
    def test_phrase_variants(self, phrase: str, price: str, should_pass: bool):
        """Parametrized test for phrase acceptance/rejection."""
        text = f"MRP {price} ({phrase})" if phrase else f"MRP {price}"
        field = _make_field(text)
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        result = RULE.check([field], req)
        if should_pass:
            assert result is None, f"Expected pass for: '{text}'"
        else:
            assert result is not None, f"Expected fail for: '{text}'"


class TestMRPRuleNumericPrice:
    """Tests for numeric price detection."""

    def test_mrp_phrase_but_no_price(self):
        """MRP has phrase but no numeric price — violation."""
        field = _make_field("MRP (Incl. of all taxes)")
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        result = RULE.check([field], req)
        assert result is not None
        assert "numeric" in result.description.lower() or "price" in result.description.lower()

    def test_mrp_with_currency_symbol(self):
        """MRP with ₹ symbol and phrase — compliant."""
        field = _make_field("₹499/- (Incl. of all taxes)")
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        result = RULE.check([field], req)
        assert result is None

    def test_mrp_with_thousands_separator(self):
        """MRP with comma-separated thousands — compliant."""
        field = _make_field("MRP Rs. 1,299/- (Incl. of all taxes)")
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        result = RULE.check([field], req)
        assert result is None


class TestMRPRuleMetadata:
    """Tests for rule metadata."""

    def test_rule_id(self):
        assert RULE.rule_id == "MRP_001"

    def test_clause(self):
        assert RULE.clause == "Rule 6(1)(e)"

    def test_severity(self):
        assert RULE.severity == "major"
