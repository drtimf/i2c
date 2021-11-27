all: build docker

build:
	env GOOS=linux GOARCH=arm go build .

docker:
	docker.exe buildx build --platform linux/arm/v7 -t drtimf/i2c:latest .
	docker login
	docker push drtimf/i2c:latest

