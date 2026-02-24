# CLASSIFICATION: UNCLASSIFIED
.PHONY: build deploy test clean

WASM_OUT = alert-validator.wasm

build:
	GOWORK=off GOOS=wasip1 GOARCH=wasm go build -o $(WASM_OUT) .

deploy: build
	rpk transform deploy $(WASM_OUT) \
		--name alert-schema-validator \
		--input-topic 'alerts.anomaly.critical,alerts.anomaly.elevated,alerts.anomaly.watch' \
		--brokers $(RTSA_REDPANDA_BROKERS)

test:
	GOWORK=off go test -v -race -count=1 ./...

clean:
	rm -f $(WASM_OUT)
