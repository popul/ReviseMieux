#!/usr/bin/env python3
"""Generate draft JSON files for a benchmark case using an LLM vision API.

Usage:
    python3 scripts/generate_json.py --case 10_SVT_cours_louis --target input
    python3 scripts/generate_json.py --case 10_SVT_cours_louis --target golden --model claude-sonnet-4-20250514
    python3 scripts/generate_json.py --case 12_physique --target input --provider openrouter --model qwen/qwen3-vl-32b-instruct
"""

import argparse
import base64
import glob
import json
import os
import sys
from pathlib import Path
import ssl
from urllib.request import Request, urlopen
from urllib.error import HTTPError

# Build SSL context using certifi certificates (fixes macOS Python SSL issues)
try:
    import certifi
    _ssl_ctx = ssl.create_default_context(cafile=certifi.where())
except ImportError:
    _ssl_ctx = None


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


PROVIDERS = {
    "anthropic": {
        "url": "https://api.anthropic.com/v1/messages",
        "env_key": "ANTHROPIC_API_KEY",
    },
    "openrouter": {
        "url": "https://openrouter.ai/api/v1/chat/completions",
        "env_key": "OPENROUTER_API_KEY",
    },
    "openai": {
        "url": "https://api.openai.com/v1/chat/completions",
        "env_key": "OPENAI_API_KEY",
    },
    "google": {
        "url": "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
        "env_key": "GOOGLE_AI_API_KEY",
    },
    "mistral": {
        "url": "https://api.mistral.ai/v1/chat/completions",
        "env_key": "MISTRAL_API_KEY",
    },
}


def call_api(content: list[dict], model: str, max_tokens: int, provider: str = "anthropic") -> str:
    """Call an LLM API and return the text response."""
    if provider == "anthropic":
        return _call_anthropic(content, model, max_tokens)
    else:
        return _call_openai_compat(content, model, max_tokens, provider)


def _call_anthropic(content: list[dict], model: str, max_tokens: int) -> str:
    """Call the Anthropic Messages API."""
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
        with urlopen(req, context=_ssl_ctx) as resp:
            body = json.loads(resp.read())
    except HTTPError as e:
        error_body = e.read().decode()
        print(f"API error {e.code}: {error_body}", file=sys.stderr)
        sys.exit(1)

    if body.get("type") == "error":
        print(f"API error: {body['error']['message']}", file=sys.stderr)
        sys.exit(1)

    return body["content"][0]["text"]


def _call_openai_compat(content: list[dict], model: str, max_tokens: int, provider: str) -> str:
    """Call an OpenAI-compatible API (OpenRouter, OpenAI, Mistral, Google)."""
    cfg = PROVIDERS[provider]
    api_key = os.environ.get(cfg["env_key"])
    if not api_key:
        print(f"Error: {cfg['env_key']} not set", file=sys.stderr)
        sys.exit(1)

    # Convert Anthropic content blocks to OpenAI format
    openai_content = []
    for block in content:
        if block["type"] == "image":
            src = block["source"]
            openai_content.append({
                "type": "image_url",
                "image_url": {"url": f"data:{src['media_type']};base64,{src['data']}"},
            })
        elif block["type"] == "text":
            openai_content.append({"type": "text", "text": block["text"]})

    payload = json.dumps({
        "model": model,
        "max_tokens": max_tokens,
        "temperature": 0,
        "messages": [{"role": "user", "content": openai_content}],
    }).encode()

    req = Request(
        cfg["url"],
        data=payload,
        headers={
            "content-type": "application/json",
            "authorization": f"Bearer {api_key}",
        },
        method="POST",
    )

    try:
        with urlopen(req, context=_ssl_ctx, timeout=120) as resp:
            body = json.loads(resp.read())
    except HTTPError as e:
        error_body = e.read().decode()
        print(f"API error {e.code}: {error_body[:500]}", file=sys.stderr)
        sys.exit(1)

    if not body.get("choices"):
        print(f"API error: no choices in response", file=sys.stderr)
        sys.exit(1)

    return body["choices"][0]["message"]["content"]


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


def dump_yaml(data: dict, f) -> None:
    """Write a dict as YAML to a file object. Stdlib only, no pyyaml needed."""
    _yaml_write(data, f, indent=0)


def _yaml_write(obj, f, indent: int) -> None:
    prefix = "  " * indent
    if isinstance(obj, dict):
        for key, val in obj.items():
            if isinstance(val, (dict, list)) and val:
                f.write(f"{prefix}{key}:\n")
                _yaml_write(val, f, indent + 1)
            else:
                _yaml_write_kv(f, prefix, key, val)
    elif isinstance(obj, list):
        for item in obj:
            if isinstance(item, dict):
                # First key on the "- " line, rest indented
                keys = list(item.keys())
                _yaml_write_kv(f, prefix + "- ", keys[0], item[keys[0]], continuation_prefix=prefix + "  ")
                for key in keys[1:]:
                    val = item[key]
                    if isinstance(val, list) and val:
                        f.write(f"{prefix}  {key}:\n")
                        for sub in val:
                            _yaml_write_kv(f, prefix + "    - ", "", sub, bare=True)
                    else:
                        _yaml_write_kv(f, prefix + "  ", key, val)
            else:
                _yaml_write_kv(f, prefix + "- ", "", item, bare=True)


def _yaml_write_kv(f, prefix: str, key: str, val, continuation_prefix: str = None, bare: bool = False) -> None:
    """Write a key: value pair, using literal block style for multiline strings."""
    s = str(val) if val is not None else ""
    if isinstance(val, str) and "\n" in s:
        # Use YAML literal block scalar (|)
        if bare:
            f.write(f"{prefix}|\n")
        else:
            f.write(f"{prefix}{key}: |\n")
        block_prefix = continuation_prefix or (prefix + "  ")
        if bare:
            block_prefix = prefix + "  "
        for line in s.split("\n"):
            f.write(f"{block_prefix}{line}\n")
    else:
        if bare:
            f.write(f"{prefix}{_yaml_scalar(val)}\n")
        else:
            f.write(f"{prefix}{key}: {_yaml_scalar(val)}\n")


def _yaml_scalar(val) -> str:
    if val is None:
        return "null"
    if isinstance(val, bool):
        return "true" if val else "false"
    if isinstance(val, (int, float)):
        return str(val)
    if isinstance(val, list) and not val:
        return "[]"
    s = str(val)
    # Quote strings that could be ambiguous or contain special chars
    # Multiline strings are handled by _yaml_write_kv with literal blocks,
    # so _yaml_scalar only handles single-line values.
    if (not s or s in ("true", "false", "null", "yes", "no")
            or s[0] in "\"'{[&*#?|->!%@`,"
            or ":" in s
            or s.startswith("- ")
            or (s[0].isdigit() and not isinstance(val, (int, float)))):
        return '"' + s.replace("\\", "\\\\").replace('"', '\\"') + '"'
    return s


# --- Prompts ---
# Single source of truth: prompts are loaded from the shared .txt files
# used by both the Go backend (via //go:embed) and this script.

_PROMPTS_DIR = os.path.join(
    os.path.dirname(__file__), "..", "..", "..", "internal", "infra", "llm", "prompts"
)


def _load_prompt(name: str) -> str:
    """Load a prompt from the shared prompts directory."""
    path = os.path.join(_PROMPTS_DIR, name)
    try:
        return Path(path).read_text()
    except FileNotFoundError:
        print(f"Error: shared prompt not found: {path}", file=sys.stderr)
        print("  Prompts live in backend/internal/infra/anthropic/prompts/", file=sys.stderr)
        sys.exit(1)


PROMPT_INPUT = _load_prompt("ocr_system.txt")

# The structuration prompt is used as-is for golden output generation.
# We wrap it with benchmark-specific context (input.json content) at call time.
_STRUCTURATION_PROMPT = _load_prompt("structuration_system.txt")

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
    parser = argparse.ArgumentParser(description="Generate benchmark JSON via LLM API")
    parser.add_argument("--case", required=True, help="Case directory name")
    parser.add_argument("--target", required=True, choices=["input", "golden", "metadata"],
                        help="Which JSON file to generate")
    parser.add_argument("--provider", default="anthropic",
                        choices=list(PROVIDERS.keys()),
                        help="API provider (default: anthropic)")
    parser.add_argument("--model", default="claude-sonnet-4-20250514", help="Model ID")
    parser.add_argument("--max-tokens", type=int, default=8192, help="Max tokens")
    parser.add_argument("--cases-dir", default="cases", help="Cases base directory")
    args = parser.parse_args()

    case_dir = os.path.join(args.cases_dir, args.case)
    images_dir = os.path.join(case_dir, "images")

    # Load images (required for input, optional for golden/metadata)
    image_blocks = []
    if os.path.isdir(images_dir):
        image_blocks = load_images(images_dir)
        print(f"  Loaded {len(image_blocks)} image(s)", file=sys.stderr)

    # Build prompt
    need_images = False
    if args.target == "input":
        if not image_blocks:
            print("Error: no images found for input generation", file=sys.stderr)
            sys.exit(1)
        need_images = True
        prompt = PROMPT_INPUT
        output_file = os.path.join(case_dir, "input.yaml")

    elif args.target == "golden":
        input_content = read_file_or_empty(os.path.join(case_dir, "input.yaml"))
        input_context = ""
        if input_content:
            input_context = f"\n\nTranscription déjà réalisée (input.yaml) :\n{input_content}"
        else:
            print("  Warning: input.yaml not found, generating from images only", file=sys.stderr)
            need_images = True
        prompt = _STRUCTURATION_PROMPT + input_context
        output_file = os.path.join(case_dir, "golden_output.yaml")

    elif args.target == "metadata":
        from datetime import date
        golden_content = read_file_or_empty(os.path.join(case_dir, "golden_output.yaml"))
        golden_context = ""
        if golden_content:
            golden_context = f"\n\nItems extraits (golden_output.yaml) :\n{golden_content}"
        pages = len(image_blocks) if image_blocks else 0
        prompt = PROMPT_METADATA.format(
            case_id=args.case,
            today=date.today().isoformat(),
            pages=pages,
            golden_context=golden_context,
        )
        output_file = os.path.join(case_dir, "metadata.yaml")

    # Build content and call API
    content = (image_blocks if need_images else []) + [{"type": "text", "text": prompt}]
    print(f"  Calling {args.provider}/{args.model}...", file=sys.stderr)
    raw = call_api(content, args.model, args.max_tokens, args.provider)

    # Parse LLM JSON response, then write as YAML
    parsed = extract_json(raw)
    with open(output_file, "w") as f:
        dump_yaml(parsed, f)

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
