"""
tests/test_applicability.py

Unit tests for applicability.py — determine_applicable_rules() and is_bulk_exempt().
Tests cover: bulk exemption, e-commerce carve-outs, food advisory, import flag.
"""

from __future__ import annotations

import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import pytest
from schemas import ValidationRequest
from applicability import determine_applicable_rules, is_bulk_exempt
from tests.test_fixtures import (
    FIXTURE_COMPLIANT_RETAIL,
    FIXTURE_ECOMMERCE_NO_MFG_DATE,
    FIXTURE_EXEMPT_BULK,
    FIXTURE_EXEMPT_CEMENT_BULK,
    FIXTURE_CEMENT_40KG_NOT_EXEMPT,
    FIXTURE_FOOD_ADVISORY,
    FIXTURE_IMPORTED_MISSING_COUNTRY,
    FIXTURE_DOMESTIC_NO_COUNTRY_REQUIRED,
    FIXTURE_EXEMPT_VOLUME,
)


def _req(fixture: dict) -> ValidationRequest:
    return ValidationRequest(**fixture)


class TestBulkExemption:
    """Tests for is_bulk_exempt()."""

    def test_exempt_over_25kg_default(self):
        """30 kg other category — exempt (> 25 kg default threshold)."""
        req = _req(FIXTURE_EXEMPT_BULK)
        exempt, reason = is_bulk_exempt(req)
        assert exempt is True
        assert "exempt_bulk_package" in reason
        assert "30" in reason or "25" in reason

    def test_not_exempt_under_25kg(self):
        """0.5 kg — not exempt."""
        req = _req(FIXTURE_COMPLIANT_RETAIL)
        exempt, reason = is_bulk_exempt(req)
        assert exempt is False
        assert reason == ""

    def test_exempt_cement_over_50kg(self):
        """55 kg cement — exempt (> 50 kg cement threshold)."""
        req = _req(FIXTURE_EXEMPT_CEMENT_BULK)
        exempt, reason = is_bulk_exempt(req)
        assert exempt is True
        assert "50" in reason

    def test_not_exempt_cement_40kg(self):
        """40 kg cement — NOT exempt (< 50 kg cement threshold)."""
        req = _req(FIXTURE_CEMENT_40KG_NOT_EXEMPT)
        exempt, reason = is_bulk_exempt(req)
        assert exempt is False

    def test_no_weight_provided(self):
        """No net_weight_kg — never exempt."""
        fixture = dict(FIXTURE_COMPLIANT_RETAIL)
        fixture["net_weight_kg"] = None
        req = _req(fixture)
        exempt, _ = is_bulk_exempt(req)
        assert exempt is False

    def test_exempt_by_volume_over_25l(self):
        """30 L other category (no weight given) — exempt by volume."""
        req = _req(FIXTURE_EXEMPT_VOLUME)
        exempt, reason = is_bulk_exempt(req)
        assert exempt is True
        assert "exempt_bulk_package" in reason
        assert "30" in reason or "L" in reason

    def test_not_exempt_under_25l(self):
        """20 L — not exempt (below 25 L threshold)."""
        fixture = dict(FIXTURE_EXEMPT_VOLUME)
        fixture["net_volume_l"] = 20.0
        req = _req(fixture)
        exempt, _ = is_bulk_exempt(req)
        assert exempt is False

    def test_both_weight_and_volume_none(self):
        """Neither weight nor volume provided — never exempt."""
        fixture = dict(FIXTURE_COMPLIANT_RETAIL)
        fixture["net_weight_kg"] = None
        req = _req(fixture)
        exempt, _ = is_bulk_exempt(req)
        assert exempt is False


class TestEcommerceApplicability:
    """
    Tests for Rule 6(10) 2017 Amendment (e-commerce).

    Rule 6(10) requires online LISTINGS to display all mandatory declarations.
    It ADDS an obligation for the platform — it does NOT remove any declaration
    from the physical label. mfg_date is mandatory on the physical label for
    ALL sale channels, including ecommerce.
    """

    def test_ecommerce_mfg_date_still_applies(self):
        """E-commerce listing — mfg_date rule MUST still apply on physical label."""
        req = _req(FIXTURE_ECOMMERCE_NO_MFG_DATE)
        rules = determine_applicable_rules(req)
        assert rules["mfg_date"]["apply"] is True, (
            "Rule 6(10) adds an online display obligation — it does NOT exempt "
            "the physical label from the mfg_date declaration. "
            "mfg_date must apply for e-commerce products."
        )

    def test_ecommerce_mfg_date_reason_is_standard(self):
        """mfg_date applicability reason for ecommerce should be 'standard mandatory rule'."""
        req = _req(FIXTURE_ECOMMERCE_NO_MFG_DATE)
        rules = determine_applicable_rules(req)
        assert "standard" in rules["mfg_date"]["reason"].lower(), (
            f"Expected 'standard' in reason, got: {rules['mfg_date']['reason']}"
        )

    def test_ecommerce_other_rules_still_apply(self):
        """E-commerce listing — all mandatory rules still apply."""
        req = _req(FIXTURE_ECOMMERCE_NO_MFG_DATE)
        rules = determine_applicable_rules(req)
        for key in ["manufacturer", "net_quantity", "mrp", "consumer_care", "mfg_date"]:
            assert rules[key]["apply"] is True, f"Expected {key} to apply for ecommerce"

    def test_retail_mfg_date_applies(self):
        """Retail listing — mfg_date rule applies."""
        req = _req(FIXTURE_COMPLIANT_RETAIL)
        rules = determine_applicable_rules(req)
        assert rules["mfg_date"]["apply"] is True


class TestFoodCategoryApplicability:
    """Tests for food category advisory downgrade."""

    def test_food_mfg_date_advisory(self):
        """Food category — mfg_date should be advisory."""
        req = _req(FIXTURE_FOOD_ADVISORY)
        rules = determine_applicable_rules(req)
        assert rules["mfg_date"]["apply"] is True
        assert rules["mfg_date"]["severity_override"] == "advisory"

    def test_food_net_quantity_advisory(self):
        """Food category — net_quantity should be advisory."""
        req = _req(FIXTURE_FOOD_ADVISORY)
        rules = determine_applicable_rules(req)
        assert rules["net_quantity"]["apply"] is True
        assert rules["net_quantity"]["severity_override"] == "advisory"

    def test_non_food_no_advisory(self):
        """Non-food category — no advisory override."""
        req = _req(FIXTURE_COMPLIANT_RETAIL)
        rules = determine_applicable_rules(req)
        assert rules["mfg_date"].get("severity_override") is None
        assert rules["net_quantity"].get("severity_override") is None


class TestImportApplicability:
    """Tests for country_of_origin rule conditional application."""

    def test_imported_product_country_rule_applies(self):
        """Imported product — country_of_origin rule must apply."""
        req = _req(FIXTURE_IMPORTED_MISSING_COUNTRY)
        rules = determine_applicable_rules(req)
        assert rules["country_of_origin"]["apply"] is True

    def test_domestic_product_country_rule_excluded(self):
        """Domestic product — country_of_origin rule must NOT apply."""
        req = _req(FIXTURE_DOMESTIC_NO_COUNTRY_REQUIRED)
        rules = determine_applicable_rules(req)
        assert rules["country_of_origin"]["apply"] is False
        assert "domestic" in rules["country_of_origin"]["reason"].lower()

    def test_standard_rules_all_apply_retail_domestic(self):
        """Standard retail domestic product — all standard rules apply."""
        req = _req(FIXTURE_COMPLIANT_RETAIL)
        rules = determine_applicable_rules(req)
        for key in ["manufacturer", "net_quantity", "mrp", "mfg_date", "consumer_care",
                    "font_size", "placement"]:
            assert rules[key]["apply"] is True, f"Rule '{key}' should apply for standard retail"
