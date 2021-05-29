
database:
	docker-compose up -d postgres
	sleep 30
	docker cp ./schema.sql grill-logger_postgres_1:/schema.sql
	docker exec -it -u postgres grill-logger_postgres_1 psql grill postgres -f /schema.sql
	docker-compose stop postgres

build:
	docker-compose build

run:
	docker-compose up -d postgres
	sleep 5
	docker-compose up
