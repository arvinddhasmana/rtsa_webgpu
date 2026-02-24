# CLASSIFICATION: UNCLASSIFIED
.PHONY: build deploy test clean

WASM_OUT = sensor-validator.wasm

build:
	GOWORK=off GOOS=wasip1 GOARCH=wasm go build -o $(WASM_OUT) .

deploy: build
	rpk transform deploy $(WASM_OUT) \
		--name sensor-schema-validator \
		--input-topic 'sensors.radar.tracks,sensors.ew.intercepts,sensors.elint.detections,sensors.isr.observations,sensors.ais.positions,sensors.cyber.iocs' \
		--brokers $(RTSA_REDPANDA_BROKERS)

test:
	GOWORK=off go test -v -race -count=1 ./...

clean:
	rm -f $(WASM_OUT)
