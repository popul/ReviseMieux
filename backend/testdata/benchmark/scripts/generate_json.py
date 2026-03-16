#!/usr/bin/env python3
"""Generate draft JSON files for a benchmark case using Claude vision API.

Usage:
    python3 scripts/generate_json.py --case 10_SVT_cours_louis --target input
    python3 scripts/generate_json.py --case 10_SVT_cours_louis --target golden
    python3 scripts/generate_json.py --case 10_SVT_cours_louis --target metadata
"""

import argparse
import base64
import glob
import json
import os
import sys
from pathlib import Path
from urllib.request import Request, urlopen
from urllib.error import HTTPError


def load_images(images_dir: str) -> list[dict]:
    """Encode all JPEG images in a directory as base64 API content blocks."""
    paths = sorted(
        glob.glob(os.path.join(images_dir, "*.jpeg"))
        + glob.glob(os.path.join(images_dir, "*.jpg"))
    )
    if not paths:
        print(f"Error: no .jpeg/.jpg files in {images_dir}", file=sys.stderr)
        sys.exit(1)

    blocks = []
    for p in paths:
        with open(p, "rb") as f:
            data = base64.b64encode(f.read()).decode()
        blocks.append({
            "type": "image",
            "source": {"type": "base64", "media_type": "image/jpeg", "data": data},
        })
    return blocks


def read_file_or_empty(path: str) -> str:
    """Read a file and return its content, or empty string if missing."""
    try:
        return Path(path).read_text()
    except FileNotFoundError:
        return ""


def call_api(content: list[dict], model: str, max_tokens: int) -> str:
    """Call the Anthropic Messages API and return the text response."""
    api_key = os.environ.get("ANTHROPIC_API_KEY")
    if not api_key:
        print("Error: ANTHROPIC_API_KEY not set", file=sys.stderr)
        sys.exit(1)

    payload = json.dumps({
        "model": model,
        "max_tokens": max_tokens,
        "messages": [{"role": "user", "content": content}],
    }).encode()

    req = Request(
        "https://api.anthropic.com/v1/messages",
        data=payload,
        headers={
            "content-type": "application/json",
            "x-api-key": api_key,
            "anthropic-version": "2023-06-01",
        },
        method="POST",
    )

    try:
        with urlopen(req) as resp:
            body = json.loads(resp.read())
    except HTTPError as e:
        error_body = e.read().decode()
        print(f"API error {e.code}: {error_body}", file=sys.stderr)
        sys.exit(1)

    if body.get("type") == "error":
        print(f"API error: {body['error']['message']}", file=sys.stderr)
        sys.exit(1)

    return body["content"][0]["text"]


def extract_json(text: str) -> dict:
    """Extract JSON from API response, stripping markdown fences if present."""
    cleaned = text.strip()
    if cleaned.startswith("```"):
        # Remove first line (```json) and last line (```)
        lines = cleaned.split("\n")
        lines = lines[1:]  # drop opening fence
        if lines and lines[-1].strip() == "```":
            lines = lines[:-1]
        cleaned = "\n".join(lines)
    return json.loads(cleaned)


# --- Prompts ---

PROMPT_INPUT = """\
Tu es un expert en transcription de cahiers scolaires. Analyse ces photos de cahier et produis une transcription structurée en blocs.

Règles :
- Transcris mot à mot, y compris les fautes.
- Chaque unité de sens = un bloc séparé.
- Types de blocs : TEXT (texte, titres, questions, réponses), DIAGRAM (schémas, graphiques), TABLE (tableaux).
- Pour les éléments visuels, décris entre crochets : [Schéma : description détaillée]
- Préserve la structure (numéros, tirets, retours à la ligne).
- Mots illisibles : [illisible]
- confidence: 1.0 pour tous les blocs.

Réponds UNIQUEMENT avec le JSON valide, sans markdown, sans commentaire :
{
  "blocks": [
    {"text": "...", "block_type": "TEXT|DIAGRAM|TABLE", "confidence": 1.0}
  ]
}"""

PROMPT_GOLDEN = """\
Tu es un expert pédagogique. À partir de ces photos de cahier scolaire, extrais les items de révision.

Types d'items :
- KNOWLEDGE : fait, définition, règle (champs: type, term, keywords, notion)
- PROCEDURE : méthode, protocole (champs: type, term, keywords, steps, notion)
- DOCUMENT : schéma/graphique à savoir lire (champs: type, term, keywords, notion)

Règles :
- 3 à 5 keywords par item
- notion = concept chapeau (2-4 notions par chapitre)
- steps uniquement pour PROCEDURE (3-6 étapes)
- Vise 6-10 items
- notions en bas = liste des notions uniques
{input_context}

Réponds UNIQUEMENT avec le JSON valide, sans markdown :
{{
  "items": [{{"type": "...", "term": "...", "keywords": [...], "notion": "..."}}],
  "notions": ["Notion1", "Notion2"]
}}"""

PROMPT_METADATA = """\
Analyse ces photos de cahier scolaire et génère le fichier metadata.json.

Champs à remplir :
- id: "{case_id}"
- subject: déduis la matière des photos
- level: classe (6e, 5e, 4e, 3e)
- topic: titre du chapitre
- difficulty: low/medium/high (pour la difficulté OCR)
- has_images: true
- expected_item_count: nombre d'items dans golden_output si disponible
- expected_types: décompte par type (KNOWLEDGE, PROCEDURE, DOCUMENT)
- ocr_source: "human_expert"
- ocr_author: "Claude Opus 4.6 (draft)"
- ocr_date: "{today}"
- ocr_metadata.handwriting_quality: soigné/normal/brouillon
- ocr_metadata.ink_types: [blue_pen, red_pen, black_pen, pencil...]
- ocr_metadata.photo_quality: good/average/poor
- ocr_metadata.content_types: [text, diagram, table, formula, map]
- ocr_metadata.support: pure_notebook/notebook_with_printout/printout_annotated
- ocr_metadata.pages: {pages}
- notes: description libre
{golden_context}

Réponds UNIQUEMENT avec le JSON valide, sans markdown."""


def main():
    parser = argparse.ArgumentParser(description="Generate benchmark JSON via Claude API")
    parser.add_argument("--case", required=True, help="Case directory name")
    parser.add_argument("--target", required=True, choices=["input", "golden", "metadata"],
                        help="Which JSON file to generate")
    parser.add_argument("--model", default="claude-opus-4-6-20250219", help="Model ID")
    parser.add_argument("--max-tokens", type=int, default=8192, help="Max tokens")
    parser.add_argument("--cases-dir", default="cases", help="Cases base directory")
    args = parser.parse_args()

    case_dir = os.path.join(args.cases_dir, args.case)
    images_dir = os.path.join(case_dir, "images")

    if not os.path.isdir(images_dir):
        print(f"Error: {images_dir} not found", file=sys.stderr)
        sys.exit(1)

    # Load images
    image_blocks = load_images(images_dir)
    print(f"  Loaded {len(image_blocks)} image(s)", file=sys.stderr)

    # Build prompt
    if args.target == "input":
        prompt = PROMPT_INPUT
        output_file = os.path.join(case_dir, "input.json")

    elif args.target == "golden":
        input_content = read_file_or_empty(os.path.join(case_dir, "input.json"))
        input_context = ""
        if input_content:
            input_context = f"\n\nTranscription déjà réalisée (input.json) :\n{input_content}"
        else:
            print("  Warning: input.json not found, generating from images only", file=sys.stderr)
        prompt = PROMPT_GOLDEN.format(input_context=input_context)
        output_file = os.path.join(case_dir, "golden_output.json")

    elif args.target == "metadata":
        from datetime import date
        golden_content = read_file_or_empty(os.path.join(case_dir, "golden_output.json"))
        golden_context = ""
        if golden_content:
            golden_context = f"\n\nItems extraits (golden_output.json) :\n{golden_content}"
        pages = len(image_blocks)
        prompt = PROMPT_METADATA.format(
            case_id=args.case,
            today=date.today().isoformat(),
            pages=pages,
            golden_context=golden_context,
        )
        output_file = os.path.join(case_dir, "metadata.json")

    # Build content and call API
    content = image_blocks + [{"type": "text", "text": prompt}]
    print(f"  Calling {args.model}...", file=sys.stderr)
    raw = call_api(content, args.model, args.max_tokens)

    # Parse and pretty-print
    parsed = extract_json(raw)
    with open(output_file, "w") as f:
        json.dump(parsed, f, ensure_ascii=False, indent=2)
        f.write("\n")

    print(f"  Written to {output_file}", file=sys.stderr)

    # Summary
    if args.target == "input":
        count = len(parsed.get("blocks", []))
        print(f"  -> {count} blocks", file=sys.stderr)
    elif args.target == "golden":
        items = len(parsed.get("items", []))
        notions = len(parsed.get("notions", []))
        print(f"  -> {items} items, {notions} notions", file=sys.stderr)


if __name__ == "__main__":
    main()
