FROM golang:1.19.2 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -a -installsuffix cgo -o app .

FROM alpine

LABEL com.dk.label-schema.name="astra-epg-aggregator" \
    com.dk.label-schema.vcs-url="https://github.com/d-kononov/epg-aggregator" \
    com.dk.label-schema.image-url="freeman1988/epg-aggregator:latest"

COPY --from=builder /app/app app

ENTRYPOINT /app/app
