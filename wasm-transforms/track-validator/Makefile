# CLASSIFICATION: UNCLASSIFIED
.PHONY: build deploy test clean

WASM_OUT = track-validator.wasm

build:
	GOWORK=off GOOS=wasip1 GOARCH=wasm go build -o $(WASM_OUT) .

deploy: build
	rpk transform deploy $(WASM_OUT) \
		--name track-schema-validator \
		--input-topic 'tracks.fused.surface,tracks.fused.air,tracks.fused.subsurface,tracks.fused.land,tracks.fused.cyber' \
		--brokers $(RTSA_REDPANDA_BROKERS)

test:
	GOWORK=off go test -v -race -count=1 ./...

clean:
	rm -f $(WASM_OUT)
