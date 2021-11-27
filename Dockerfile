FROM alpine:latest

COPY Orbitron-Medium.ttf ./Orbitron-Medium.ttf
COPY i2c ./i2c

CMD ./i2c
