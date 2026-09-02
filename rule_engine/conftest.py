"""
conftest.py — pytest configuration for the legal-metrology-rule-engine test suite.

Adds the rule_engine root to sys.path so tests can import from the top-level modules
without package installation.
"""

import sys
import os

# Add rule_engine root directory to path
sys.path.insert(0, os.path.dirname(__file__))
