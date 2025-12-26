#!/usr/bin/env python3
"""
OpenAPI YAML to JSON converter with validation.
Validates the YAML structure and converts to properly formatted JSON.
"""

import json
import sys
import yaml
from pathlib import Path


def validate_openapi_structure(spec: dict) -> list[str]:
    """
    Validate basic OpenAPI 3.x structure.
    Returns list of validation errors (empty if valid).
    """
    errors = []

    # Check required top-level fields
    if "openapi" not in spec:
        errors.append("Missing required field: openapi")
    elif not spec["openapi"].startswith("3."):
        errors.append(f"Invalid OpenAPI version: {spec['openapi']} (expected 3.x)")

    if "info" not in spec:
        errors.append("Missing required field: info")
    else:
        info = spec["info"]
        if "title" not in info:
            errors.append("Missing required field: info.title")
        if "version" not in info:
            errors.append("Missing required field: info.version")

    if "paths" not in spec:
        errors.append("Missing required field: paths")

    # Validate components references
    if "components" in spec:
        components = spec["components"]
        defined_schemas = set()
        defined_params = set()
        defined_security = set()

        if "schemas" in components:
            defined_schemas = set(components["schemas"].keys())
        if "parameters" in components:
            defined_params = set(components["parameters"].keys())
        if "securitySchemes" in components:
            defined_security = set(components["securitySchemes"].keys())

        # Check for dangling references in paths
        if "paths" in spec:
            refs_found = []
            find_refs(spec["paths"], refs_found)

            for ref in refs_found:
                if ref.startswith("#/components/schemas/"):
                    schema_name = ref.split("/")[-1]
                    if schema_name not in defined_schemas:
                        errors.append(f"Undefined schema reference: {ref}")
                elif ref.startswith("#/components/parameters/"):
                    param_name = ref.split("/")[-1]
                    if param_name not in defined_params:
                        errors.append(f"Undefined parameter reference: {ref}")

    return errors


def find_refs(obj, refs: list, path=""):
    """Recursively find all $ref values in the spec."""
    if isinstance(obj, dict):
        for key, value in obj.items():
            if key == "$ref" and isinstance(value, str):
                refs.append(value)
            else:
                find_refs(value, refs, f"{path}.{key}")
    elif isinstance(obj, list):
        for i, item in enumerate(obj):
            find_refs(item, refs, f"{path}[{i}]")


def count_elements(spec: dict) -> dict:
    """Count various elements in the spec for reporting."""
    counts = {
        "paths": 0,
        "operations": 0,
        "schemas": 0,
        "parameters": 0,
        "tags": set(),
    }

    if "paths" in spec:
        counts["paths"] = len(spec["paths"])
        for path, methods in spec["paths"].items():
            if isinstance(methods, dict):
                for method, operation in methods.items():
                    if method.lower() in [
                        "get",
                        "post",
                        "put",
                        "patch",
                        "delete",
                        "options",
                        "head",
                    ]:
                        counts["operations"] += 1
                        if isinstance(operation, dict) and "tags" in operation:
                            for tag in operation["tags"]:
                                counts["tags"].add(tag)

    if "components" in spec:
        if "schemas" in spec["components"]:
            counts["schemas"] = len(spec["components"]["schemas"])
        if "parameters" in spec["components"]:
            counts["parameters"] = len(spec["components"]["parameters"])

    counts["tags"] = len(counts["tags"])
    return counts


def main():
    docs_dir = Path(__file__).parent
    yaml_file = docs_dir / "openapi.yaml"
    json_file = docs_dir / "openapi.json"

    print("=" * 60)
    print("OpenAPI Validation and JSON Generation")
    print("=" * 60)

    # Load YAML
    print(f"\n[1/4] Loading YAML from: {yaml_file}")
    try:
        with open(yaml_file, "r", encoding="utf-8") as f:
            spec = yaml.safe_load(f)
        print("      YAML loaded successfully")
    except yaml.YAMLError as e:
        print("      ERROR: Invalid YAML syntax")
        print(f"      {e}")
        sys.exit(1)
    except FileNotFoundError:
        print(f"      ERROR: File not found: {yaml_file}")
        sys.exit(1)

    # Validate structure
    print("\n[2/4] Validating OpenAPI structure...")
    errors = validate_openapi_structure(spec)

    if errors:
        print(f"      Found {len(errors)} validation error(s):")
        for error in errors:
            print(f"      - {error}")
        sys.exit(1)
    else:
        print("      Structure validation passed")

    # Count elements
    print("\n[3/4] Analyzing specification...")
    counts = count_elements(spec)
    print(f"      - OpenAPI version: {spec.get('openapi', 'unknown')}")
    print(f"      - Title: {spec.get('info', {}).get('title', 'unknown')}")
    print(f"      - Version: {spec.get('info', {}).get('version', 'unknown')}")
    print(f"      - Paths: {counts['paths']}")
    print(f"      - Operations: {counts['operations']}")
    print(f"      - Schemas: {counts['schemas']}")
    print(f"      - Parameters: {counts['parameters']}")
    print(f"      - Tags: {counts['tags']}")

    # Write JSON
    print(f"\n[4/4] Writing JSON to: {json_file}")
    try:
        with open(json_file, "w", encoding="utf-8") as f:
            json.dump(spec, f, indent=2, ensure_ascii=False)
        print("      JSON written successfully")

        # Report file sizes
        yaml_size = yaml_file.stat().st_size / 1024
        json_size = json_file.stat().st_size / 1024
        print("\n      File sizes:")
        print(f"      - YAML: {yaml_size:.1f} KB")
        print(f"      - JSON: {json_size:.1f} KB")

    except IOError as e:
        print("      ERROR: Failed to write JSON")
        print(f"      {e}")
        sys.exit(1)

    print("\n" + "=" * 60)
    print("Validation and generation completed successfully!")
    print("=" * 60)

    return 0


if __name__ == "__main__":
    sys.exit(main())
