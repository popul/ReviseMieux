#!/bin/bash
# Wrapper that calls the shared report script
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
exec "${REPO_ROOT}/.claude/scripts/report-execution.sh" "sprint"
