.PHONY: test docker docker_stop remove restart purge_restart

test:
	docker compose up db -d
	cd account_service && go test -v ./tests/...

docker: test
	docker compose up -d --build

docker_stop:
	docker compose down

remove: docker_stop
	docker volume rm transaction-system_pg_storage
	docker volume rm transaction-system_rabbitmq_storage
	docker volume rm transaction-system_rabbitmq
	docker image rm account_service
	docker image rm transaction_service

restart: docker_stop docker

purge_restart: remove docker