# syntax=docker/dockerfile:1

FROM golang:1.21.5 AS builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o interface-alert-report

# Use distroless as the base image for a smaller footprint
FROM gcr.io/distroless/base

# Set the working directory inside the distroless image
WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /app/interface-alert-report /

# Expose the port the application will run on
EXPOSE 8080

# Run the application
CMD ["/interface-alert-report"]
