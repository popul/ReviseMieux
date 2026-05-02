"""Revise Mieux web service.

Async job pattern to avoid Cloudflare's 100s edge timeout: the POST /jobs
endpoint returns immediately with a job ID, while the LLM call runs in the
background. The frontend polls GET /jobs/{id} every few seconds and downloads
the resulting HTML via GET /jobs/{id}/result when done.
"""
import asyncio
import base64
import json
import logging
import os
import sys
import time
from enum import Enum
from pathlib import Path
from typing import Optional
from uuid import uuid4

import httpx
from fastapi import BackgroundTasks, FastAPI, File, Form, HTTPException, UploadFile
from fastapi.responses import HTMLResponse, JSONResponse, Response
from jinja2 import Environment, FileSystemLoader, select_autoescape
from pydantic import BaseModel, Field, ValidationError

from schema import Fiche

LLM_URL = os.environ.get("LLM_URL", "http://macbook-pro-2.home:1234/v1/chat/completions")
LLM_MODEL = os.environ.get("LLM_MODEL", "qwen/qwen3.6-35b-a3b")
PROMPT_PATH = Path(os.environ.get("PROMPT_PATH", "/code/prompt.txt"))
INDEX_PATH = Path(os.environ.get("INDEX_PATH", "/code/index.html"))
SHELL_PATH = Path(os.environ.get("SHELL_PATH", "/code/shell.html.j2"))
LLM_TIMEOUT = float(os.environ.get("LLM_TIMEOUT", "600"))
JOB_TTL_SECONDS = float(os.environ.get("JOB_TTL_SECONDS", "3600"))
LOG_LEVEL = os.environ.get("LOG_LEVEL", "INFO").upper()

logging.basicConfig(
    level=LOG_LEVEL,
    format="%(asctime)s %(levelname)s %(name)s %(message)s",
    stream=sys.stdout,
    force=True,
)
log = logging.getLogger("revise")

SYSTEM_PROMPT = PROMPT_PATH.read_text(encoding="utf-8")
INDEX_HTML = INDEX_PATH.read_text(encoding="utf-8")

# Jinja env trusts |safe-marked HTML emitted by the LLM. Autoescape is enabled
# for plain `{{ value }}` substitutions so unmarked text is still escaped.
JINJA_ENV = Environment(
    loader=FileSystemLoader(str(SHELL_PATH.parent)),
    autoescape=select_autoescape(default=False),
    trim_blocks=False,
    lstrip_blocks=False,
)
SHELL_TEMPLATE = JINJA_ENV.get_template(SHELL_PATH.name)
log.info(
    "boot llm_url=%s llm_model=%s timeout=%ss prompt_chars=%d index_chars=%d shell=%s",
    LLM_URL, LLM_MODEL, LLM_TIMEOUT, len(SYSTEM_PROMPT), len(INDEX_HTML), SHELL_PATH.name,
)


class JobStatus(str, Enum):
    pending = "pending"
    running = "running"
    done = "done"
    error = "error"


class Job(BaseModel):
    id: str
    status: JobStatus = JobStatus.pending
    error: Optional[str] = None
    created_at: float
    finished_at: Optional[float] = None
    # html stored separately to keep status responses small
    model_config = {"extra": "ignore"}


class GenerationContext(BaseModel):
    subject: str = Field(default="", description="Matière et niveau, ex: 'PC 4eme'")
    target: str = Field(default="", description="Note visée, ex: '15-17'")
    extra: str = Field(default="", description="Précisions libres")


JOBS: dict[str, Job] = {}
JOB_HTML: dict[str, str] = {}

app = FastAPI(title="Revise Mieux Web")


def _gc_jobs() -> None:
    """Drop jobs older than TTL to keep memory bounded."""
    now = time.time()
    for jid in list(JOBS):
        if now - JOBS[jid].created_at > JOB_TTL_SECONDS:
            JOBS.pop(jid, None)
            JOB_HTML.pop(jid, None)


def _target_band(target: str) -> str:
    t = (target or "").lower()
    if any(s in t for s in ("18", "19", "20", "excellen")):
        return "excellence"
    if any(s in t for s in ("15", "16", "17", "bonne")):
        return "good"
    if any(s in t for s in ("12", "13", "14", "moyenne", "correct")):
        return "basic"
    return "good"


_BAND_LABELS = {
    "basic": "12-14 (Sections 1→6, ni 8 ni 9)",
    "good": "15-17 (Sections 1→6 + Section 8 partielle a/b/c, pas de Section 9)",
    "excellence": "18-20 (Sections 1→6 + Section 8 a→g complète + Section 9 OBLIGATOIRE)",
}

# Compact AC summary, 1 short line per criterion. The full spec lives in
# prompt.txt under "## ✅ Critères d'acceptance"; this is the salience
# reminder appended at the end of the user message so the model has the
# must-haves under its eyes when it stops generating.
_AC_REMINDERS_COMMON = [
    "[AC-CONT-02] UNE mini-carte SVG par notion (3-7 feuilles), pas une mindmap géante.",
    "[AC-CONT-03] ≥3 documents inédits Section 3 : ≥1 tableau de données + ≥1 graphique/schéma SVG + ≥1 doc adapté à la matière.",
    "[AC-CONT-04] Si HG/SVT/français/PC : encadré « MÉTHODE — Analyser un document (N-D-S-I-D) » présent en Section 3F.",
    "[AC-CONT-05] Si TRAP identifiés : encadré « PIÈGES CLASSIQUES » avec tableau Confusion / Distinction / Mnémo (≥3 entrées).",
    "[AC-CONT-06] Si HG/français : ≥3 exercices complets d'analyse de document (1 texte + 1 iconographique + 1 cartographique) avec corrigé MODÈLE RÉDIGÉ + erreurs à éviter — section dédiée 3H bis.",
    "[AC-CONT-07] Session 1 : exactement 15 Q numérotées Q1→Q15, ≥1 MINDMAP_PARTIAL.",
    "[AC-CONT-08] Session 2 : exactement 15 Q numérotées Q16→Q30, interleaving, ≥1 MINDMAP_RECALL.",
    "[AC-CONT-09] CHAQUE corrigé S1+S2 (et S3 si présente, en abrégé) : 3 lignes EXACTES « ✅ Réponse attendue : … » / « 🔍 Ce qui manque souvent : … » / « 💡 Astuce mnémonique : … ». Le label texte « Astuce mnémonique » est OBLIGATOIRE, pas seulement l'emoji 💡.",
    "[AC-CONT-11] Tableau de progression mastery présent : colonnes Item | Type | Notion | Après S1 | Après S2 | Prochaine révision.",
    "[AC-HTML-04] DEUX boutons d'impression fixed top-right OBLIGATOIRES : « 🖨️ Version élève » (ajoute classe `print-eleve`) ET « 🖨️ Version parent » (ajoute classe `print-parent`). Un seul = non conforme.",
    "[AC-HTML-05] Menu hamburger + drawer latéral avec backdrop, masqué à l'impression.",
]
_AC_REMINDERS_GOOD = [
    "[AC-15-01] Section 8 partielle (a) frise + (b) vocabulaire exhaustif + (c) pièges renforcés.",
]
_AC_REMINDERS_EXCELLENCE = [
    "[AC-20-01] Section 8 COMPLÈTE — TOUTES les sous-parties (a) frise + (b) vocabulaire exhaustif + (c) pièges renforcés + (d) fiches personnages (si ≥2 personnages dans le cours) + (e) tableau comparatif des notions proches + (f) rédactions modèles + grille d'évaluation + (g) méthodes par type de document. Aucune ne doit être omise.",
    "[AC-20-04] Section 9 (Session 3) : 10 questions numérotées Q31→Q40 (continuité S1+S2), JAMAIS Q1-Q10.",
    "[AC-20-05] La question MINDMAP_REBUILD de la Section 9 est un SVG inline ~500×300 px (thème central + 4-6 branches vides + feuilles vides + corrigé dans <details>). ASCII art ou texte plat « 1.____ » = non conforme.",
    "[AC-20-06] Section 9 reprend ≥1 TRAP et ≥1 SHORT_LIST identifiés à l'Étape 2.",
    "[AC-20-07] Corrigé Session 3 abrégé présent (1-2 lignes par question).",
]


def _checklist_for_target(target: str) -> str:
    band = _target_band(target)
    items = list(_AC_REMINDERS_COMMON)
    if band == "good":
        items += _AC_REMINDERS_GOOD
    elif band == "excellence":
        items += _AC_REMINDERS_EXCELLENCE
    header = (
        "⚠️ RAPPEL DES CRITÈRES D'ACCEPTANCE BLOQUANTS — la spec complète vit dans la "
        "section « ✅ Critères d'acceptance » du prompt système. Ci-dessous les AC les "
        "plus souvent oubliés : AVANT de fermer </body>, vérifie chacun et ajoute la "
        "pièce manquante si besoin.\n\n"
        f"Cible détectée : **{_BAND_LABELS[band]}**.\n\n"
    )
    body = "\n".join(items)
    footer = (
        "\n\nTu ne produis QUE du HTML (pas de prose hors balises, pas de bloc ```html, "
        "pas de description de ta démarche)."
    )
    return header + body + footer


def _build_user_content(ctx: GenerationContext, image_data_urls: list[str]) -> list:
    parts = ["Voici les photos d'un cours/cahier."]
    if ctx.subject:
        parts.append(f"Matière/niveau : {ctx.subject}.")
    if ctx.target:
        parts.append(f"Note visée : {ctx.target}.")
    if ctx.extra:
        parts.append(f"Précisions : {ctx.extra}.")
    parts.append(_checklist_for_target(ctx.target))
    parts.append("Produis directement le HTML complet de la fiche de révision (rien d'autre).")
    content: list = [{"type": "text", "text": "\n\n".join(parts)}]
    for url in image_data_urls:
        content.append({"type": "image_url", "image_url": {"url": url}})
    return content


def _strip_html_wrapper(html: str) -> str:
    if html.startswith("```"):
        html = html.split("\n", 1)[1] if "\n" in html else html
        if html.endswith("```"):
            html = html.rsplit("```", 1)[0]
    return html.strip()


def _strip_thinking(text: str) -> str:
    # LM Studio returns reasoning in a separate `reasoning_content` field, but if
    # the server ever inlines it, drop any leading <think>...</think> block.
    while text.lstrip().startswith("<think>"):
        end = text.find("</think>")
        if end == -1:
            break
        text = text[end + len("</think>") :]
    return text.lstrip()


async def _run_llm(job_id: str, ctx: GenerationContext, image_urls: list[str]) -> str:
    payload = {
        "model": LLM_MODEL,
        "messages": [
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": _build_user_content(ctx, image_urls)},
        ],
        "temperature": 0.4,
        "max_tokens": 32000,
        "chat_template_kwargs": {"enable_thinking": True},
    }
    img_bytes = sum(len(u) for u in image_urls)
    user_text = next(
        (p["text"] for p in payload["messages"][1]["content"] if p.get("type") == "text"),
        "",
    )
    log.info(
        "job=%s llm_request model=%s images=%d image_payload_bytes=%d "
        "system_chars=%d user_text_chars=%d target_band=%s max_tokens=%d temperature=%s",
        job_id, LLM_MODEL, len(image_urls), img_bytes,
        len(SYSTEM_PROMPT), len(user_text), _target_band(ctx.target),
        payload["max_tokens"], payload["temperature"],
    )
    t0 = time.monotonic()
    async with httpx.AsyncClient(timeout=LLM_TIMEOUT) as client:
        r = await client.post(LLM_URL, json=payload)
    elapsed = time.monotonic() - t0
    if r.status_code != 200:
        log.error(
            "job=%s llm_http_error status=%d elapsed=%.1fs body=%s",
            job_id, r.status_code, elapsed, r.text[:500],
        )
        raise RuntimeError(f"LLM HTTP {r.status_code}: {r.text[:500]}")
    data = r.json()
    choice = (data.get("choices") or [{}])[0]
    msg = choice.get("message") or {}
    finish = choice.get("finish_reason")
    usage = data.get("usage") or {}
    reasoning_tokens = (usage.get("completion_tokens_details") or {}).get("reasoning_tokens")
    reasoning_chars = len(msg.get("reasoning_content") or "")
    log.info(
        "job=%s llm_response elapsed=%.1fs finish=%s prompt_tokens=%s "
        "completion_tokens=%s reasoning_tokens=%s reasoning_chars=%d",
        job_id, elapsed, finish, usage.get("prompt_tokens"),
        usage.get("completion_tokens"), reasoning_tokens, reasoning_chars,
    )
    if finish == "length":
        log.warning(
            "job=%s output_truncated finish_reason=length — bump max_tokens or shorten prompt",
            job_id,
        )
    try:
        html = msg["content"]
    except KeyError as exc:
        log.error("job=%s unexpected_response keys=%s", job_id, list(msg.keys()))
        raise RuntimeError(f"unexpected LLM response: {data}") from exc
    raw_len = len(html or "")
    html = _strip_thinking(html or "")
    html = _strip_html_wrapper(html)
    ends_with_html = html.rstrip().lower().endswith("</html>")
    log.info(
        "job=%s html raw_chars=%d clean_chars=%d ends_with_html=%s",
        job_id, raw_len, len(html), ends_with_html,
    )
    if not ends_with_html:
        log.warning(
            "job=%s html_likely_truncated last80=%r", job_id, html[-80:],
        )
    if not html.lower().startswith("<!doctype") and "<html" not in html.lower():
        log.error("job=%s html_invalid first80=%r", job_id, html[:80])
        raise RuntimeError("LLM did not return HTML")
    return html


async def _process_job(job_id: str, ctx: GenerationContext, image_urls: list[str]) -> None:
    job = JOBS[job_id]
    job.status = JobStatus.running
    log.info("job=%s status=running images=%d", job_id, len(image_urls))
    t0 = time.monotonic()
    try:
        html = await _run_llm(job_id, ctx, image_urls)
        JOB_HTML[job_id] = html
        job.status = JobStatus.done
        log.info(
            "job=%s status=done elapsed=%.1fs html_chars=%d",
            job_id, time.monotonic() - t0, len(html),
        )
    except Exception as exc:
        job.status = JobStatus.error
        job.error = str(exc)
        log.exception(
            "job=%s status=error elapsed=%.1fs error=%s",
            job_id, time.monotonic() - t0, exc,
        )
    finally:
        job.finished_at = time.time()


FIXTURE_PATH = Path(os.environ.get("FIXTURE_PATH", "/code/fixture.json"))


def _render_fiche(fiche: Fiche) -> str:
    return SHELL_TEMPLATE.render(**fiche.model_dump(by_alias=True))


@app.get("/", response_class=HTMLResponse)
async def home() -> HTMLResponse:
    return HTMLResponse(INDEX_HTML)


@app.get("/healthz")
async def healthz() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/preview", response_class=HTMLResponse)
async def preview() -> HTMLResponse:
    """Render the bundled fixture through the shell — no LLM call.
    Used to validate the chrome (drawer, print buttons, theme, mobile, print mode)
    without spending tokens. Visit /preview in any browser to test."""
    raw = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))
    fiche = Fiche.model_validate(raw)
    return HTMLResponse(_render_fiche(fiche))


@app.get("/preview/raw")
async def preview_raw() -> JSONResponse:
    """Return the fixture as JSON — useful for inspecting the schema shape."""
    return JSONResponse(json.loads(FIXTURE_PATH.read_text(encoding="utf-8")))


@app.post("/jobs")
async def create_job(
    background: BackgroundTasks,
    files: list[UploadFile] = File(...),
    subject: str = Form(""),
    target: str = Form(""),
    extra: str = Form(""),
) -> JSONResponse:
    _gc_jobs()
    if not files:
        raise HTTPException(400, "no files")

    image_urls: list[str] = []
    raw_bytes = 0
    file_summary: list[str] = []
    for f in files:
        data = await f.read()
        if not data:
            continue
        mime = f.content_type or "image/jpeg"
        b64 = base64.b64encode(data).decode("ascii")
        image_urls.append(f"data:{mime};base64,{b64}")
        raw_bytes += len(data)
        file_summary.append(f"{f.filename or '?'}({mime},{len(data)}B)")

    if not image_urls:
        raise HTTPException(400, "no readable images")

    ctx = GenerationContext(subject=subject, target=target, extra=extra)
    job = Job(id=str(uuid4()), created_at=time.time())
    JOBS[job.id] = job
    log.info(
        "job=%s created images=%d raw_bytes=%d subject=%r target=%r extra_chars=%d files=%s",
        job.id, len(image_urls), raw_bytes, ctx.subject, ctx.target,
        len(ctx.extra), ",".join(file_summary)[:300],
    )
    asyncio.create_task(_process_job(job.id, ctx, image_urls))
    return JSONResponse({"job_id": job.id})


@app.get("/jobs/{job_id}")
async def job_status(job_id: str) -> JSONResponse:
    job = JOBS.get(job_id)
    if not job:
        raise HTTPException(404, "job not found")
    payload = {
        "id": job.id,
        "status": job.status.value,
        "error": job.error,
        "created_at": job.created_at,
        "finished_at": job.finished_at,
    }
    return JSONResponse(payload)


@app.get("/jobs/{job_id}/result")
async def job_result(job_id: str) -> Response:
    job = JOBS.get(job_id)
    if not job:
        raise HTTPException(404, "job not found")
    if job.status != JobStatus.done:
        raise HTTPException(409, f"job not done: {job.status.value}")
    html = JOB_HTML.get(job_id)
    if html is None:
        raise HTTPException(500, "job done but no result available")
    return Response(
        content=html,
        media_type="text/html; charset=utf-8",
        headers={"content-disposition": 'attachment; filename="fiche-revision.html"'},
    )
