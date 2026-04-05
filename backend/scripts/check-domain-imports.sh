#!/usr/bin/env bash
# check-domain-imports.sh — Verifies that domain/ packages never import
# infra/, http/, app/, or external framework packages.
# This is a compile-time-free architectural guard.

set -euo pipefail

RED='\033[31m'
GREEN='\033[32m'
RESET='\033[0m'

DOMAIN_DIR="internal/domain"
VIOLATIONS=0

# Forbidden import patterns for domain packages
FORBIDDEN_PATTERNS=(
    '"github.com/gin-gonic'
    '"github.com/jackc/pgx'
    '"github.com/go-redis'
    '"github.com/minio'
    '"github.com/anthropics'
    'internal/infra'
    'internal/http'
    'internal/app'
    'internal/config'
    'internal/db'
)

for gofile in $(find "$DOMAIN_DIR" -name '*.go' -not -name '*_test.go'); do
    for pattern in "${FORBIDDEN_PATTERNS[@]}"; do
        if grep -qn "$pattern" "$gofile" 2>/dev/null; then
            line=$(grep -n "$pattern" "$gofile" | head -1)
            echo -e "${RED}VIOLATION${RESET}: $gofile imports forbidden package"
            echo "  $line"
            VIOLATIONS=$((VIOLATIONS + 1))
        fi
    done
done

if [ "$VIOLATIONS" -gt 0 ]; then
    echo ""
    echo -e "${RED}$VIOLATIONS import violation(s) found in domain/${RESET}"
    echo "Domain packages must not import infrastructure, HTTP, or app packages."
    exit 1
else
    echo -e "${GREEN}Domain import check passed${RESET} — no violations"
    exit 0
fi
