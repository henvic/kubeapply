.SILENT: server processor notifier test
.PHONY: server processor notifier test
server:
	go run ./cmd/server
processor:
	go run ./cmd/processor
notifier:
	go run ./cmd/notifier
test:
	./scripts/test.sh
