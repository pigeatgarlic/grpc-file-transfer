db_name = maxex
postgres_user = root
postgres_password = secret
postgres_port = 5432
postgres_address = localhost

download-postgres:
	docker pull postgres

postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=$(postgres_user) -e POSTGRES_PASSWORD=$(postgres_password) -d postgres

createdb:
	docker exec -it postgres createdb --username=root --owner=root $(db_name)

dropdb:
	docker exec -it postgres dropdb $(db_name)

migrateup:
	migrate -path db/migration -database postgresql://$(postgres_user):$(postgres_password)@$(postgres_address):$(postgres_port)/$(db_name)?sslmode=disable -verbose up

migratedown:
	migrate -path db/migration -database postgresql://$(postgres_user):$(postgres_password)@$(postgres_address):$(postgres_port)/$(db_name)?sslmode=disable -verbose down

.PHONY: download-postgres postgres createdb dropdb migrateup migratedown