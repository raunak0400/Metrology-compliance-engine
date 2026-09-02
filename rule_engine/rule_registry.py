"""
rule_registry.py — Central registry of all active rules + version metadata.

This is the single file to edit when:
  - Adding a new rule (import and register it here)
  - Retiring a rule (remove its entry)
  - Viewing which ruleset version is active

The ruleset_version string is read from rules_config.json and identifies exactly
which version of the Legal Metrology (Packaged Commodities) Rules is being enforced.
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import Dict, List

from rules.manufacturer_rule import ManufacturerRule
from rules.net_quantity_rule import NetQuantityRule
from rules.mrp_rule import MRPRule
from rules.mfg_date_rule import MfgDateRule
from rules.consumer_care_rule import ConsumerCareRule
from rules.country_of_origin_rule import CountryOfOriginRule
from rules.font_size_rule import FontSizeRule
from rules.placement_rule import PlacementRule
from rules.generic_name_rule import GenericNameRule
from rules.base import Rule

# ---------------------------------------------------------------------------
# Load version metadata from central config
# ---------------------------------------------------------------------------
_CONFIG_PATH = Path(__file__).parent / "rules_config.json"
with _CONFIG_PATH.open(encoding="utf-8") as _f:
    _CONFIG = json.load(_f)

RULESET_VERSION: str = _CONFIG["ruleset_version"]
RULESET_DESCRIPTION: str = _CONFIG["ruleset_description"]

# ---------------------------------------------------------------------------
# Rule instances (stateless singletons — rules are pure functions)
# ---------------------------------------------------------------------------

_MANUFACTURER_RULE = ManufacturerRule()
_NET_QUANTITY_RULE = NetQuantityRule()
_MRP_RULE = MRPRule()
_MFG_DATE_RULE = MfgDateRule()
_CONSUMER_CARE_RULE = ConsumerCareRule()
_COUNTRY_ORIGIN_RULE = CountryOfOriginRule()
_FONT_SIZE_RULE = FontSizeRule()
_PLACEMENT_RULE = PlacementRule()
_GENERIC_NAME_RULE = GenericNameRule()

# ---------------------------------------------------------------------------
# Registry: maps rule_key -> Rule instance
# The rule_key must match keys used in applicability.py's determine_applicable_rules()
# ---------------------------------------------------------------------------
RULE_REGISTRY: Dict[str, Rule] = {
    "manufacturer": _MANUFACTURER_RULE,
    "net_quantity": _NET_QUANTITY_RULE,
    "mrp": _MRP_RULE,
    "mfg_date": _MFG_DATE_RULE,
    "consumer_care": _CONSUMER_CARE_RULE,
    "country_of_origin": _COUNTRY_ORIGIN_RULE,
    "font_size": _FONT_SIZE_RULE,
    "placement": _PLACEMENT_RULE,
    "generic_name": _GENERIC_NAME_RULE,
}

# ---------------------------------------------------------------------------
# Multi-violation rules: these rules can return a list of violations.
# The engine handles them specially via check_all().
# ---------------------------------------------------------------------------
MULTI_VIOLATION_RULES = frozenset(["font_size", "placement"])


def get_all_rule_metadata() -> List[dict]:
    """
    Return metadata for all registered rules — used by GET /rules endpoint.
    """
    metadata = []
    for rule_key, rule_instance in RULE_REGISTRY.items():
        metadata.append(
            {
                "rule_id": rule_instance.rule_id,
                "clause": rule_instance.clause,
                "severity": rule_instance.severity,
                "rule_key": rule_key,
                "rule_class": type(rule_instance).__name__,
                "description": _CONFIG["rules_metadata"].get(
                    f"{rule_key}_rule", {}
                ).get("description", "See rule class docstring"),
            }
        )
    return metadata
