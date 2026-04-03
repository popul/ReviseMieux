#!/bin/bash
# .claude/scripts/report-execution.sh
# Called by the Stop hook of each skill via ${CLAUDE_SKILL_DIR}/scripts/generate-report.sh

COMMAND_NAME="$1"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%S")
REPORT_DIR=".claude/reports/${COMMAND_NAME}"
REPORT_FILE="${REPORT_DIR}/${TIMESTAMP}.md"

mkdir -p "${REPORT_DIR}"

cat > "${REPORT_FILE}" << EOF
---
command: ${COMMAND_NAME}
execution_id: ${COMMAND_NAME}_${TIMESTAMP}
started_at: ${TIMESTAMP}
finished_at: $(date -u +"%Y-%m-%dT%H:%M:%S")
status: pending
---

<!-- Rapport à compléter par Claude via les instructions du skill -->
EOF

# Count reports for this cycle (exclude INDEX.md and CRITERIA.md)
COUNT=$(ls -1 "${REPORT_DIR}"/*.md 2>/dev/null | grep -cv 'INDEX\|CRITERIA' | tr -d ' ')

# If threshold reached, create flag
if [ "$COUNT" -ge 5 ]; then
  touch "${REPORT_DIR}/.review-ready"
fi

echo "📝 Rapport #${COUNT} créé pour /${COMMAND_NAME}"
