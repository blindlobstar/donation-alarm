.PHONY: test integration

test:
	go test --tags="unit" ./...

integration:
	docker compose -f docker-compose.test.yml up backend --build --abort-on-container-exit
	docker compose -f docker-compose.test.yml down --volumes