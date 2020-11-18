prep:
	docker kill db || true
	docker rm db || true
	docker run -d --name db -e POSTGRES_PASSWORD=password -p 5432:5432 postgres:alpine
	sleep 5
	docker exec -ti db createdb -U postgres db

clean:
	docker kill db || true
	docker rm db || true
