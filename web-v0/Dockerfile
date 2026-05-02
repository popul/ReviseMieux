# Build context = repo root (not web-v0/) so we can pull
# prompts/study-guide/system.md as the canonical prompt source.
# Cf. .github/workflows/web-v0.yml.
FROM python:3.12-slim

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

WORKDIR /app

COPY web-v0/requirements.txt /app/
RUN pip install --no-cache-dir -r requirements.txt

# Application code (web-v0/) and the shared prompt source (prompts/study-guide/).
COPY web-v0/main.py web-v0/schema.py web-v0/index.html web-v0/shell.html.j2 web-v0/fixture.json /app/
COPY prompts/study-guide/system.md /app/system.md

ENV PROMPT_PATH=/app/system.md \
    INDEX_PATH=/app/index.html \
    SHELL_PATH=/app/shell.html.j2 \
    FIXTURE_PATH=/app/fixture.json

EXPOSE 8000

# Drop privileges. /app contents are read-only at runtime.
RUN useradd -r -u 1000 -g 0 -s /sbin/nologin app && chown -R app:0 /app
USER 1000

HEALTHCHECK --interval=15s --timeout=5s --start-period=20s \
    CMD python -c "import urllib.request,sys; sys.exit(0 if urllib.request.urlopen('http://localhost:8000/healthz',timeout=3).status==200 else 1)" || exit 1

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
