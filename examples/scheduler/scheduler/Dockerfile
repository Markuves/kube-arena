FROM golang:1.22-bookworm AS builder

WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /scheduler ./...

FROM gcr.io/distroless/base-debian12
COPY --from=builder /scheduler /scheduler

ENTRYPOINT ["/scheduler"]

