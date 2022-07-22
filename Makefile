start:
	docker-compose up --build

restart: cleanup start

cleanup:
	docker-compose down --volumes --remove-orphans