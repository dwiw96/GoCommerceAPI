pg-run:
	sudo docker compose -f ./pkg/driver/postgresql/docker-compose-pg.yml up -d
pg-start:
	sudo docker container start pg_vocagame_test
pg-exec:
	sudo docker exec -it pg_vocagame_test psql -p 5432 -h localhost -U dwiw -d technical_test
pg-stop:
	sudo docker container stop pg_vocagame_test

rd-run:
	sudo docker run --name rd_vocagame_test -p 6379:6379 -d redis
rd-start:
	sudo docker container start rd_vocagame_test
rd-exec:
	sudo docker exec -it rd_vocagame_test redis-cli -h localhost
rd-stop:
	sudo docker container stop rd_vocagame_test

migrate-create:
	migrate create -ext sql -dir internal/migrations -seq init
migrate-up:
	migrate -path internal/migrations -database "postgresql://dwiw:secret@localhost:5432/technical_test?sslmode=disable" -verbose up $(v)
migrate-down:
	migrate -path internal/migrations -database "postgresql://dwiw:secret@localhost:5432/technical_test?sslmode=disable" -verbose down $(v)
migrate-force:
	migrate -path internal/migrations -database "postgresql://dwiw:secret@localhost:5432/technical_test?sslmode=disable" force $(v)
