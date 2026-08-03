# Stage 1: Build the Go application

# We use a multi-stage build to keep the final image small and secure. 
# The first stage uses the official Golang image to compile the application
# while the second stage uses a minimal Alpine Linux image to run the compiled binary.
FROM golang:1.24.6-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Cache layer for Go modules to speed up builds. This layer will only be rebuilt if go.mod or go.sum changes.
COPY go.sum go.mod ./
RUN go mod download

# Copy the source code into the container
COPY . .

# CGO_ENABLED=0 disables CGO, which is necessary for building a statically linked binary. This is important for running the binary in a minimal container without the need for additional shared libraries.
# it makes the binary more portable and reduces the size of the final image.
# GOOS=linux ensures the binary is built for Linux, which is the target OS for our container.
RUN CGO_ENABLED=0 GOOS=linux go build -o pong-service ./cmd/main.go

# Stage 2: Create a minimal image for the application
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy the compiled binary from the builder stage to the final image. This keeps the final image small and secure, as it only contains the necessary binary and its dependencies.
COPY --from=builder /app/pong-service .

# if any vulnerability is found in the base image, we can use a non-root user to decrease potential risks
USER nonroot:nonroot

EXPOSE 8081

ENTRYPOINT ["./pong-service"]
