pg-run:
	sudo docker compose -f ./pkg/driver/postgresql/docker-compose-pg.yml up -d
pg-start:
	sudo docker container start pg_GoCommerceApi
pg-exec:
	sudo docker exec -it pg_GoCommerceApi psql -p 5432 -h localhost -U dwiw -d commerce_main_db
pg-stop:
	sudo docker container stop pg_GoCommerceApi

rd-run:
	sudo docker run --name rd_GoCommerceAPI -p 6379:6379 -d redis
rd-start:
	sudo docker container start rd_GoCommerceAPI
rd-exec:
	sudo docker exec -it rd_GoCommerceAPI redis-cli -h localhost
rd-stop:
	sudo docker container stop rd_GoCommerceAPI

migrate-create:
	migrate create -ext sql -dir internal/migrations -seq init
migrate-up:
	migrate -path internal/migrations -database "postgresql://dwiw:secret@localhost:5432/commerce_main_db?sslmode=disable" -verbose up $(v)
migrate-down:
	migrate -path internal/migrations -database "postgresql://dwiw:secret@localhost:5432/commerce_main_db?sslmode=disable" -verbose down $(v)
migrate-force:
	migrate -path internal/migrations -database "postgresql://dwiw:secret@localhost:5432/commerce_main_db?sslmode=disable" force $(v)
