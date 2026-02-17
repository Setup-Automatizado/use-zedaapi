#!/usr/bin/env python3
"""Generate openapi.json from openapi.yaml (single source of truth)."""

import json
import sys
from pathlib import Path

try:
    import yaml
except ImportError:
    print("ERROR: PyYAML not installed. Run: pip install pyyaml", file=sys.stderr)
    sys.exit(1)

SCRIPT_DIR = Path(__file__).parent
YAML_PATH = SCRIPT_DIR / "openapi.yaml"
JSON_PATH = SCRIPT_DIR / "openapi.json"


def main():
    if not YAML_PATH.exists():
        print(f"ERROR: {YAML_PATH} not found", file=sys.stderr)
        sys.exit(1)

    with open(YAML_PATH, encoding="utf-8") as f:
        spec = yaml.safe_load(f)

    with open(JSON_PATH, "w", encoding="utf-8") as f:
        json.dump(spec, f, indent=2, ensure_ascii=False)
        f.write("\n")

    yaml_paths = len(spec.get("paths", {}))
    yaml_schemas = len(spec.get("components", {}).get("schemas", {}))
    print(f"Generated {JSON_PATH.name} from {YAML_PATH.name}")
    print(f"  Paths: {yaml_paths}, Schemas: {yaml_schemas}")


if __name__ == "__main__":
    main()
