FROM golang:1.16-alpine as builder

ENV GO111MODULE=on

COPY go.mod /go/src/cosmos-exporter/
COPY go.sum /go/src/cosmos-exporter/
WORKDIR /go/src/cosmos-exporter/

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cosmos-exporter


FROM alpine:3.13

RUN apk add --update --no-cache ca-certificates
COPY --from=builder /go/src/cosmos-exporter/cosmos-exporter /cosmos-exporter

CMD ["/cosmos-exporter"]
