#!/usr/bin/env python3
"""Merge benchmark results from multiple runs into a consolidated summary.json.

IDP: 3 runs (main + LM Studio corrected + gemini-3.1-flash-lite-preview)
OCR: 2 runs (main + LM Studio corrected)

Logic:
- For each model, keep the entry with the highest avg_composite_score.
- Remove obsolete models from the catalog.
"""

import json
import os
import shutil

BASE = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
RESULTS = os.path.join(BASE, "testdata", "benchmark", "results")

# Models removed from catalog
IDP_OBSOLETE = {
    "claude-sonnet-4-6", "gpt-4o", "gpt-4o-mini", "o3-mini",
    "deepseek-reasoner", "gemini-2.5-pro", "mistral-large-latest",
}
OCR_OBSOLETE = {
    "claude-sonnet-4-6", "gpt-4o", "gpt-4o-mini",
    "gemini-2.5-pro", "mistral-large-latest",
}


def load_summary(path):
    with open(path) as f:
        return json.load(f)


def merge_summaries(files, obsolete):
    """Merge multiple summary files, keeping best result per model."""
    best = {}  # model -> entry

    for path in files:
        entries = load_summary(path)
        for entry in entries:
            model = entry["model"]
            if model in obsolete:
                continue
            existing = best.get(model)
            if existing is None:
                best[model] = entry
            elif entry["avg_composite_score"] > existing["avg_composite_score"]:
                best[model] = entry

    # Sort by composite score descending
    result = sorted(best.values(), key=lambda e: e["avg_composite_score"], reverse=True)
    return result


def write_consolidated(bench_type, files, obsolete):
    out_dir = os.path.join(RESULTS, bench_type, "2026-04-18_consolidated")
    os.makedirs(out_dir, exist_ok=True)

    merged = merge_summaries(files, obsolete)
    out_path = os.path.join(out_dir, "summary.json")
    with open(out_path, "w") as f:
        json.dump(merged, f, indent=2, ensure_ascii=False)

    print(f"[{bench_type.upper()}] Wrote {len(merged)} models to {out_path}")
    for entry in merged:
        print(f"  {entry['avg_composite_score']:.4f}  {entry['provider']}/{entry['model']}")


def main():
    # IDP
    idp_files = [
        os.path.join(RESULTS, "idp", "2026-03-28_06h58", "summary.json"),
        os.path.join(RESULTS, "idp", "2026-04-18_15h02", "summary.json"),
        os.path.join(RESULTS, "idp", "2026-04-18_15h04", "summary.json"),
    ]
    write_consolidated("idp", idp_files, IDP_OBSOLETE)

    print()

    # OCR
    ocr_files = [
        os.path.join(RESULTS, "ocr", "2026-03-28_07h02", "summary.json"),
        os.path.join(RESULTS, "ocr", "2026-04-18_15h14", "summary.json"),
    ]
    write_consolidated("ocr", ocr_files, OCR_OBSOLETE)


if __name__ == "__main__":
    main()
