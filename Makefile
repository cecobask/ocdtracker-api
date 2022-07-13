start:
	docker-compose up --build --detach

restart: cleanup start

cleanup:
	docker-compose down --volumes --remove-orphans