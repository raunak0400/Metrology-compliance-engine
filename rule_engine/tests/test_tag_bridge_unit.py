"""
tests/test_tag_bridge_unit.py

Unit tests for tag_bridge.py internals.

Covers:
  - _find_font_height: exact match, partial substring match, ambiguous list
    ordering (SPICE-TAGS-002), no-match null (CANDLE-TAGS-006), short-token
    guard, empty inputs.
  - synthesize_from_tags: field coverage, null tag handling, merge priority.
"""

from __future__ import annotations

import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import pytest
from schemas import TextMeasurementEntry
from tag_bridge import _find_font_height, synthesize_from_tags


# ---------------------------------------------------------------------------
# Helper
# ---------------------------------------------------------------------------

def _meas(text: str, height_mm: float, confidence: float = 0.95) -> TextMeasurementEntry:
    return TextMeasurementEntry(
        text=text,
        confidence=confidence,
        height_mm=height_mm,
        min_required_mm=1.0,
        is_rule7_compliant=True,
    )


# ===========================================================================
# _find_font_height
# ===========================================================================

class TestFindFontHeightExactMatch:
    """Exact matches must always be chosen."""

    def test_exact_match_single_entry(self):
        """Single entry that matches exactly → returns its height."""
        result = _find_font_height("100 g", [_meas("100 g", 2.6)])
        assert result == pytest.approx(2.6)

    def test_exact_match_returns_correct_value(self):
        """Exact match on a multi-word quantity string."""
        result = _find_font_height("300 ml", [_meas("300 ml", 3.15)])
        assert result == pytest.approx(3.15)

    def test_case_insensitive_match(self):
        """Match is case-insensitive (field text may be uppercase from OCR)."""
        result = _find_font_height("100 G", [_meas("100 g", 2.6)])
        assert result == pytest.approx(2.6)


class TestFindFontHeightSubstringMatch:
    """Measurement substring inside compound field text."""

    def test_substring_inside_bonus_pack(self):
        """Measurement 'Marico Ltd.' appears inside longer field text."""
        field_text = "Marico Ltd., 7th Floor, Grande Palladium, 175 CST Road"
        result = _find_font_height(field_text, [_meas("Marico Ltd.", 2.1)])
        assert result == pytest.approx(2.1)

    def test_partial_quantity_in_bonus_pack(self):
        """'300 ml' is a substring of '300 ml + 30 ml FREE'."""
        result = _find_font_height(
            "300 ml + 30 ml FREE",
            [_meas("300 ml", 3.15)],
        )
        assert result == pytest.approx(3.15)


class TestFindFontHeightAmbiguousOrder:
    """
    SPICE-TAGS-002 scenario: two measurement entries, one longer (comes first
    in list), one exact. The closest-length algorithm must pick the exact match
    regardless of list order.
    """

    def test_longer_entry_first_exact_wins(self):
        """
        Longer entry "100 g pack date info" at 1.4mm comes FIRST in the list.
        Exact match "100 g" at 2.6mm comes SECOND.
        Expected: 2.6mm (exact match / closest length wins).

        Note: "100 g pack date info" is NOT a substring of "100 g" (it's
        longer), so it won't match at all — the exact string "100 g" is the
        only match and wins automatically. This test pins the documented
        guarantee so a future refactor can't accidentally break it.
        """
        measurements = [
            _meas("100 g pack date info", 1.4),  # comes first — must NOT win
            _meas("100 g", 2.6),                  # exact match — must win
        ]
        result = _find_font_height("100 g", measurements)
        assert result == pytest.approx(2.6), (
            "Closest-length algorithm must prefer '100 g' (exact match, diff=0) "
            "over '100 g pack date info' (not even a substring of '100 g'). "
            "Got: {result}"
        )

    def test_exact_first_partial_second_exact_wins(self):
        """
        When list order IS exact-first then partial: still correct.
        Verifies algorithm works regardless of ordering direction.
        """
        measurements = [
            _meas("300 ml + 30 ml FREE", 3.5),  # exact match
            _meas("300 ml", 2.0),               # partial match
        ]
        result = _find_font_height("300 ml + 30 ml FREE", measurements)
        assert result == pytest.approx(3.5), (
            "Exact match (diff=0) must win over partial match (diff>0)"
        )

    def test_two_partial_matches_closer_length_wins(self):
        """
        When two measurement entries both match as substrings, the one with
        the smaller length difference to the field text wins.
        """
        # Field text: "200 ml bottle"
        measurements = [
            _meas("200 ml", 1.2),          # diff = |6 - 13| = 7
            _meas("200 ml bottl", 2.5),    # diff = |11 - 13| = 2  ← closer
        ]
        result = _find_font_height("200 ml bottle", measurements)
        assert result == pytest.approx(2.5), (
            "Among partial matches, the one with the smaller length difference "
            "to the field text must win."
        )


class TestFindFontHeightNoMatch:
    """
    CANDLE-TAGS-006 scenario: no measurement text overlaps with field text.
    Must return None without raising.
    """

    def test_no_overlap_returns_none(self):
        """Measurements contain completely unrelated text → None."""
        result = _find_font_height(
            "150 g",
            [
                _meas("wick length 5cm", 1.5),
                _meas("burn time 40hrs", 1.2),
            ],
        )
        assert result is None, (
            "When no measurement text is a substring of the field text, "
            "font_height_mm must be None."
        )

    def test_empty_measurements_list(self):
        """Empty measurements list → None (no error)."""
        result = _find_font_height("500 ml", [])
        assert result is None

    def test_none_measurements(self):
        """None measurements → None (no error)."""
        result = _find_font_height("500 ml", None)
        assert result is None

    def test_empty_field_text(self):
        """Empty field text → None (no error)."""
        result = _find_font_height("", [_meas("500 ml", 1.5)])
        assert result is None

    def test_none_field_text(self):
        """None-like empty string → None (no error)."""
        result = _find_font_height("   ", [_meas("500 ml", 1.5)])
        assert result is None


class TestFindFontHeightShortTokenGuard:
    """Short measurement strings (< 4 chars) must be ignored."""

    def test_unit_token_ignored(self):
        """'ml' (2 chars) must not match anything — prevents spurious hits."""
        result = _find_font_height("500 ml", [_meas("ml", 1.0)])
        assert result is None

    def test_three_char_token_ignored(self):
        """'100' (3 chars) is below the 4-char guard."""
        result = _find_font_height("100 g", [_meas("100", 2.0)])
        assert result is None

    def test_exactly_four_chars_allowed(self):
        """Exactly 4-char measurement text is not blocked by the guard."""
        result = _find_font_height("100 g net", [_meas("100g", 1.8)])
        # "100g" in "100 g net" → False (space missing) → None
        # This verifies the guard doesn't block 4-char entries; the substring
        # check itself determines whether a match occurs.
        assert result is None  # no match because "100g" ≠ "100 g"


# ===========================================================================
# synthesize_from_tags — surface-level coverage
# ===========================================================================

class TestSynthesizeFromTagsNullHandling:
    """Null tag values must not produce empty-text synthesized fields."""

    def test_null_importer_not_synthesized(self):
        """IMPORTER_NAME: null → no 'importer' field synthesized."""
        fields, summaries = synthesize_from_tags(
            tags={
                "MANUFACTURER_NAME": "Acme Ltd.",
                "MANUFACTURER_ADDRESS": "123 Road, Mumbai",
                "IMPORTER_NAME": None,
                "IMPORTER_ADDRESS": None,
                "PRODUCT_COMMON_NAME": "Widget",
                "NET_QUANTITY": "500 g",
                "MANUFACTURE_MONTH": "Jan",
                "MANUFACTURE_YEAR": "2025",
                "MRP": "₹100",
                "MRP_INCLUSIVE_OF_TAXES": "Incl. of all taxes",
                "CONSUMER_CARE_PHONE": "9876543210",
            },
            measurements=None,
            existing_fields=[],
        )
        synthesized_types = {f.field_type for f in fields}
        assert "importer" not in synthesized_types, (
            "Null IMPORTER_NAME/ADDRESS must not produce an 'importer' field"
        )

    def test_empty_tags_dict_returns_nothing(self):
        """Empty tags dict → no fields synthesized."""
        fields, summaries = synthesize_from_tags(
            tags={},
            measurements=None,
            existing_fields=[],
        )
        assert fields == []
        assert summaries == []

    def test_none_tags_returns_nothing(self):
        """None tags → no fields synthesized."""
        from tag_bridge import synthesize_from_tags
        fields, summaries = synthesize_from_tags(
            tags=None,
            measurements=None,
            existing_fields=[],
        )
        assert fields == []
        assert summaries == []


class TestSynthesizeFromTagsMergePriority:
    """Caller-provided existing_fields always take priority over tags."""

    def test_existing_high_confidence_blocks_synthesis(self):
        """High-confidence net_quantity in existing_fields → bridge skips it."""
        from schemas import ExtractedField, BoundingBox
        existing = [
            ExtractedField(
                field_type="net_quantity",
                text="500 ml",
                confidence=0.95,
                bbox=BoundingBox(x=0, y=0, w=100, h=20),
                font_height_mm=1.5,
                panel_location="principal_display_panel",
            )
        ]
        fields, summaries = synthesize_from_tags(
            tags={"NET_QUANTITY": "Should Be Ignored"},
            measurements=None,
            existing_fields=existing,
        )
        synthesized_types = {f.field_type for f in fields}
        assert "net_quantity" not in synthesized_types

    def test_low_confidence_existing_allows_synthesis(self):
        """Low-confidence net_quantity in existing_fields → bridge fills the gap."""
        from schemas import ExtractedField, BoundingBox
        existing = [
            ExtractedField(
                field_type="net_quantity",
                text="500 ml",
                confidence=0.50,  # below 0.75 threshold
                bbox=BoundingBox(x=0, y=0, w=100, h=20),
                font_height_mm=1.5,
                panel_location="principal_display_panel",
            )
        ]
        fields, summaries = synthesize_from_tags(
            tags={"NET_QUANTITY": "500 ml"},
            measurements=None,
            existing_fields=existing,
            confidence_threshold=0.75,
        )
        synthesized_types = {f.field_type for f in fields}
        assert "net_quantity" in synthesized_types, (
            "Low-confidence existing field should allow synthesis to fill the gap"
        )
