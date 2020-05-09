FROM golang:1.13-alpine as builder
ENV CGO_ENABLED=0
WORKDIR /go/src/app
COPY go.* ./
RUN go mod download && go mod verify
COPY . .

RUN go build -ldflags="-s -w"

FROM scratch
LABEL maintainer="patrick246"
COPY --from=builder /etc/ssl/cert.pem /etc/ssl/cert.pem
COPY --from=builder /go/src/app/mongodb-query-exporter /mongodb-query-exporter

ENTRYPOINT ["/mongodb-query-exporter"]
EXPOSE 9736
