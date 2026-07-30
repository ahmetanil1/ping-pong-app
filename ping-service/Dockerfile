FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.sum go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ping-service ./cmd/main.go

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/ping-service .

# Copy the templates directory from the builder stage to the final image. This is necessary for serving the HTML templates used by the ping service.
COPY --from=builder /app/web/templates ./web/templates

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["./ping-service"]