#!/bin/bash
set -e

SERVICES=("svc-ais-ingestion" "svc-cyber-ingestion" "svc-elint-ingestion" "svc-isr-ingestion")

for SVC in "${SERVICES[@]}"; do
    # 1. Update main.go
    MAIN_FILE=$(find "$SVC/cmd" -name "main.go" | head -n 1)
    if [ -f "$MAIN_FILE" ] && ! grep -q "cfg.Coverage" "$MAIN_FILE"; then
        sed -i 's/obsProd, dlqObsProd, auditEmitter, logger)/obsProd, dlqObsProd, auditEmitter, logger, cfg.Coverage)/g' "$MAIN_FILE"
        echo "Updated $MAIN_FILE"
    fi
    
    # 2. Update ingestion_test.go
    TEST_FILE="$SVC/internal/handler/ingestion_test.go"
    if [ -f "$TEST_FILE" ] && ! grep -q "logger, nil" "$TEST_FILE"; then
        sed -i 's/, logger)/, logger, nil)/g' "$TEST_FILE"
        echo "Updated $TEST_FILE"
    fi
done
