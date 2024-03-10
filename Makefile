all: multiarch

build:
	go build .

build-linux-arm:
	env GOOS=linux GOARCH=arm go build -o i2c-linux-arm .

docker:
	docker buildx build --platform linux/arm/v7 -t drtimf/i2c:latest .
	docker login
	docker push drtimf/i2c:latest

multiarch:
	docker login
	docker buildx build -f Dockerfile.multiarch --push --platform linux/amd64,linux/arm/v7,linux/arm64 -t drtimf/i2c:latest .

test: build
	./i2c

test-pi3: build-linux-arm
	scp -P 19123 i2c-linux-arm pi@192.168.0.10:

query:
	curl -s http://localhost:2112/metrics

clean:
	rm -f i2c

