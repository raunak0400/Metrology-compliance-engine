"""
applicability.py — Determines which rules apply to a given product validation request.

This module implements the conditional rule-selection logic based on:
  - Product category (food, cosmetics, electronics, cement, fertilizer, etc.)
  - Sale channel (retail vs. e-commerce)
  - Net weight (bulk exemption thresholds from Chapter II)
  - Import status (country_of_origin only for imported goods)

The output is a structured dict consumed by engine.py — the engine never hardcodes
which rules run; all that logic lives here.

Legal References:
  - Chapter II exemptions: Legal Metrology (Packaged Commodities) Rules, 2011 — Rule 3
  - Rule 6(10) 2017 Amendment: adds online-display obligation for e-commerce platforms
    (does NOT exempt any field from the physical label)
  - Food category: Prevention of Food Adulteration Act considerations
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import Dict, Optional, Tuple

from schemas import ValidationRequest

# Load bulk exemption thresholds from config (single source of truth)
_CONFIG_PATH = Path(__file__).parent / "rules_config.json"
with _CONFIG_PATH.open(encoding="utf-8") as _f:
    _CONFIG = json.load(_f)

_BULK_THRESHOLDS: dict = _CONFIG["bulk_exemption_thresholds"]
_FOOD_ADVISORY: list = _CONFIG["food_advisory_fields"]
_IMPORTED_REQUIRED: list = _CONFIG["imported_required_fields"]

# Canonical set of all rule keys — maps to the rule classes in rule_registry.py
ALL_RULE_KEYS = frozenset(
    [
        "manufacturer",
        "net_quantity",
        "mrp",
        "mfg_date",
        "consumer_care",
        "country_of_origin",
        "font_size",
        "placement",
        "generic_name",
    ]
)


def _get_bulk_threshold_kg(category: str) -> float:
    """
    Return the applicable minimum weight exemption threshold in kg.
    Defaults to the standard 25 kg threshold for unlisted categories.
    """
    return _BULK_THRESHOLDS.get(
        f"{category.lower()}_kg", _BULK_THRESHOLDS["default_kg"]
    )


def is_bulk_exempt(request: "ValidationRequest") -> Tuple[bool, str]:
    """
    Determine whether this package is exempt from Chapter II rules entirely.

    Checks BOTH weight (net_weight_kg) and volume (net_volume_l) against
    category-specific thresholds from the single 'bulk_exemption_thresholds'
    dict in rules_config.json. A package is exempt if EITHER dimension
    exceeds its threshold.

    Returns:
        (True, reason_string) if exempt, (False, "") otherwise.
    """
    category_lower = request.category.lower()

    # --- Weight-based exemption ---
    if request.net_weight_kg is not None:
        weight_threshold = _BULK_THRESHOLDS.get(
            f"{category_lower}_kg", _BULK_THRESHOLDS["default_kg"]
        )
        if request.net_weight_kg > weight_threshold:
            return True, (
                f"exempt_bulk_package: net weight {request.net_weight_kg:.2f} kg exceeds "
                f"the Chapter II exemption threshold of {weight_threshold:.0f} kg for "
                f"category '{request.category}'"
            )

    # --- Volume-based exemption ---
    if request.net_volume_l is not None:
        volume_threshold = _BULK_THRESHOLDS.get(
            f"{category_lower}_l", _BULK_THRESHOLDS["default_l"]
        )
        if request.net_volume_l > volume_threshold:
            return True, (
                f"exempt_bulk_package: net volume {request.net_volume_l:.2f} L exceeds "
                f"the Chapter II exemption threshold of {volume_threshold:.0f} L for "
                f"category '{request.category}'"
            )

    return False, ""


def determine_applicable_rules(
    request: ValidationRequest,
) -> Dict[str, dict]:
    """
    Evaluate which rules should run for this validation request.

    Returns a dict mapping rule_key -> rule_config, where rule_config contains:
        {
            "apply": bool,          # True = run this rule
            "severity_override": Optional[str],  # e.g. "advisory"
            "reason": str,          # human-readable explanation for the decision
        }

    Example output:
        {
            "manufacturer": {"apply": True,  "severity_override": None, "reason": "standard mandatory rule"},
            "mfg_date":     {"apply": True,  "severity_override": None, "reason": "standard mandatory rule"},
            "country_of_origin": {"apply": False, "severity_override": None, "reason": "domestic product"},
            ...
        }
    """
    result: Dict[str, dict] = {}

    for rule_key in ALL_RULE_KEYS:
        result[rule_key] = {
            "apply": True,
            "severity_override": None,
            "reason": "standard mandatory rule",
        }

    # -------------------------------------------------------------------------
    # Rule 1 — E-commerce: Rule 6(10) 2017 Amendment
    # Rule 6(10) requires that the ONLINE LISTING of a pre-packaged commodity
    # clearly display all mandatory declarations — it ADDS an obligation for the
    # e-commerce platform, it does NOT remove any declaration from the physical
    # label. mfg_date remains mandatory on the physical package regardless of
    # sale channel.
    # -------------------------------------------------------------------------
    # (No rules are removed for ecommerce — all physical-label checks still apply)

    # -------------------------------------------------------------------------
    # Rule 2 — Food category: downgrade certain rules to 'advisory'
    # Food articles may be governed by Prevention of Food Adulteration Act / FSSAI
    # for certain fields. Still check, but tag lower severity.
    # -------------------------------------------------------------------------
    if request.category == "food":
        for field_key in _FOOD_ADVISORY:
            rule_key = _field_to_rule_key(field_key)
            if rule_key in result and result[rule_key]["apply"]:
                result[rule_key]["severity_override"] = "advisory"
                result[rule_key]["reason"] = (
                    "Food category — this field may also be governed by FSSAI / "
                    "Prevention of Food Adulteration Act. Flagged as advisory rather "
                    "than mandatory under LM Rules for cross-regulatory harmony."
                )

    # -------------------------------------------------------------------------
    # Rule 3 — Country of origin: only applies to imported goods
    # -------------------------------------------------------------------------
    if not request.is_imported:
        result["country_of_origin"]["apply"] = False
        result["country_of_origin"]["reason"] = (
            "Domestic product — Rule 6(1)(b) proviso (country of origin) only "
            "applies to imported commodities"
        )
    else:
        result["country_of_origin"]["apply"] = True
        result["country_of_origin"]["reason"] = (
            "Imported product — Rule 6(1)(b) proviso requires country of origin declaration"
        )

    return result


def _field_to_rule_key(field_key: str) -> str:
    """Map a field_type/config key to the canonical rule key used in the result dict."""
    mapping = {
        "mfg_date": "mfg_date",
        "net_quantity": "net_quantity",
        "country_of_origin": "country_of_origin",
        "manufacturer": "manufacturer",
        "mrp": "mrp",
        "consumer_care": "consumer_care",
    }
    return mapping.get(field_key, field_key)
