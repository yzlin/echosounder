FROM golang:1.10-alpine as builder

COPY . /go/src/echosounder

WORKDIR /go/src/echosounder

RUN mkdir build && CGO_ENABLED=0 go build -o build/echosounderd -v ./cmd/echosounderd

FROM alpine

LABEL authors="yzlin <yzlin1985@gmail.com>"

RUN set -x \
    && apk add --no-cache \
    ca-certificates

COPY --from=builder /go/src/echosounder/build/echosounderd /app/

ENV PATH=/app:$PATH

VOLUME /app/etc
WORKDIR /app

EXPOSE 10023 10080

CMD ["echosounderd"]
