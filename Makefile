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

deps:
	mkdir -p assets
	( cd assets && wget https://code.jquery.com/jquery-3.7.1.min.js )
	( cd assets && wget "https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css" )
	( cd assets && wget "https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js" )

test: build
	./i2c

test-pi4: build-linux-arm
	scp -r i2c-linux-arm config pi@192.168.0.20:

query:
	curl -s http://localhost:2112/metrics

BULB_MAC=d0:73:d5:64:72:09

test-lifx:
	curl http://192.168.0.20:2112/lifx/$(BULB_MAC) | jq .
	make test-lifx-on
	sleep 1
	make test-lifx-off

test-lifx-on:
	curl -X POST -d 'power=on'  http://192.168.0.20:2112/lifx/$(BULB_MAC)

test-lifx-off:
	curl -X POST -d 'power=off'  http://192.168.0.20:2112/lifx/$(BULB_MAC)

clean:
	rm -f i2c

