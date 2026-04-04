#!/bin/bash
# Record a Maestro flow with iOS Simulator screen recording
# Usage: ./record-flow.sh flows/01-onboarding.yaml videos/01-onboarding.mp4
set -e

FLOW="${1:?Usage: $0 <flow.yaml> <output.mp4>}"
OUTPUT="${2:?Usage: $0 <flow.yaml> <output.mp4>}"

mkdir -p "$(dirname "$OUTPUT")"
rm -f "$OUTPUT"

echo "Starting screen recording..."
xcrun simctl io booted recordVideo --codec h264 "$OUTPUT" &
RECORD_PID=$!
sleep 1

echo "Running flow: $FLOW"
maestro test "$FLOW" || true

sleep 1
echo "Stopping recording..."
kill -INT $RECORD_PID 2>/dev/null
wait $RECORD_PID 2>/dev/null

if [ -f "$OUTPUT" ] && [ -s "$OUTPUT" ]; then
  echo "Video: $OUTPUT ($(du -h "$OUTPUT" | cut -f1))"
  open "$OUTPUT"
else
  echo "ERROR: Video not created"
  exit 1
fi
