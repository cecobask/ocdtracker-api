start:
	docker-compose up --build

cleanup:
	docker-compose down --volumes --remove-orphans