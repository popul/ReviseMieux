"""Pydantic schema describing the JSON the LLM is asked to produce.

Stage 1 (LLM call) returns one Fiche; Stage 2 (server) renders it through
shell.html.j2. Fields marked _html are raw HTML fragments emitted by the LLM
that flow through Jinja's |safe filter (we trust them in this controlled
internal tool — sanitize before exposing to untrusted authors).
"""
from typing import Literal, Optional

from pydantic import BaseModel, ConfigDict, Field


TargetBand = Literal["basic", "good", "excellence"]
Mastery = Literal["UNKNOWN", "FRAGILE", "OK"]


class Meta(BaseModel):
    title: str
    subject: str = ""
    target: str = ""
    target_band: TargetBand = "good"


class Notion(BaseModel):
    id: str
    name: str
    icon: str = ""
    svg: str = ""
    items: list[str] = []
    memo: str = ""


class Document(BaseModel):
    letter: str
    title: str
    type: Literal["table", "graph", "scheme", "map", "text", "stat"]
    html: str


class Trap(BaseModel):
    confusion: str
    distinction: str
    memo: str


class DocExercise(BaseModel):
    type: Literal["text", "iconographic", "cartographic", "statistical"]
    doc_html: str
    questions: list[str]
    model_answer_html: str
    errors: list[str] = []


class Question(BaseModel):
    id: str
    type: str = ""
    notion: str = ""
    stem_html: str
    options: list[str] = []
    svg: str = ""


class Corrige(BaseModel):
    qid: str
    answer_html: str
    common_miss: str
    mnemonic: str


class Plan(BaseModel):
    j0: str
    j1: str
    j2: str


class MasteryRow(BaseModel):
    item: str
    type: str
    notion: str
    after_s1: Mastery
    after_s2: Mastery
    next_review: str


class Personnage(BaseModel):
    name: str
    dates: str = ""
    role: str
    ideas: list[str] = []


class CompareRow(BaseModel):
    a: str
    b: str
    diff: str
    example: str = ""


class GridRow(BaseModel):
    criterion: str
    points: str
    indicators: str


class Redaction(BaseModel):
    prompt: str
    model_html: str
    grid: list[GridRow]


class MethodeDoc(BaseModel):
    type: str
    steps: list[str]


class VocabEntry(BaseModel):
    # `def` is a reserved word in Python, so we expose it through the alias.
    model_config = ConfigDict(populate_by_name=True)
    term: str
    definition: str = Field(alias="def")


class Annexes(BaseModel):
    frise_html: str = ""
    vocab: list[VocabEntry] = []
    traps_renforces: list[Trap] = []
    personnages: list[Personnage] = []
    comparatif: list[CompareRow] = []
    redactions: list[Redaction] = []
    methodes_par_type_doc: list[MethodeDoc] = []


class Session3CorrigeRow(BaseModel):
    qid: str
    answer: str


class Session3(BaseModel):
    questions: list[Question]
    rebuild_mindmap_svg: str = ""
    rebuild_mindmap_corrige_svg: str = ""
    corrige_abrege: list[Session3CorrigeRow] = []


class Fiche(BaseModel):
    meta: Meta
    summary: list[str]
    notions: list[Notion]
    documents: list[Document]
    method_doc: str = ""
    traps: list[Trap] = []
    doc_exercises: list[DocExercise] = []
    session1: list[Question]
    session2: list[Question]
    corriges: list[Corrige]
    plan: Plan
    mastery: list[MasteryRow]
    annexes: Optional[Annexes] = None
    session3: Optional[Session3] = None
