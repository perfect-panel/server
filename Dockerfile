# Use a smaller base image for the build stage
FROM golang:alpine AS builder

LABEL stage=gobuilder

ARG TARGETARCH
ARG VERSION
ENV CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH}

# Combine apk commands into one to reduce layer size
RUN apk update --no-cache && apk add --no-cache tzdata ca-certificates

WORKDIR /build

# Copy go.mod and go.sum first to take advantage of Docker caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the binary with version and build time
RUN BUILD_TIME=$(date -u +"%Y-%m-%d %H:%M:%S") && \
    go build -ldflags="-s -w -X 'github.com/perfect-panel/server/pkg/constant.Version=${VERSION}' -X 'github.com/perfect-panel/server/pkg/constant.BuildTime=${BUILD_TIME}'" -o /app/ppanel ppanel.go

# Final minimal image
FROM scratch

# Copy CA certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai

ENV TZ=Asia/Shanghai

# Set working directory and copy binary
WORKDIR /app

COPY --from=builder /app/ppanel /app/ppanel
COPY --from=builder /build/etc /app/etc

# Expose the port (optional)
EXPOSE 8080

# Specify entry point
ENTRYPOINT ["/app/ppanel"]
CMD ["run", "--config", "etc/ppanel.yaml"]