# Use a smaller base image for the build stage
FROM golang:alpine AS builder

LABEL stage=gobuilder

ARG TARGETARCH
ENV CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH}

# Combine apk commands into one to reduce layer size
RUN apk update --no-cache && apk add --no-cache tzdata ca-certificates

WORKDIR /build

# Copy go.mod and go.sum first to take advantage of Docker caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the binary with optimization flags to reduce binary size
RUN go build -ldflags="-s -w" -o /app/ppanel ppanel.go

# Final minimal image
FROM scratch

# Copy CA certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai

ENV TZ=Asia/Shanghai

# Set working directory and copy binary
WORKDIR /app

COPY --from=builder /app/ppanel /app/ppanel
COPY --from=builder /etc /app/etc

# Expose the port (optional)
EXPOSE 8080

# Specify entry point
ENTRYPOINT ["/app/ppanel"]
CMD ["run", "--config", "etc/ppanel.yaml"]
