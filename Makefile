start:
	docker-compose up --build --detach

cleanup:
	docker-compose down --volumes --remove-orphans