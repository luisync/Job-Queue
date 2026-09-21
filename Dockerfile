# Image of the job queue application.

FROM golang:1.27-alpine AS builder

WORKDIR /app

# Handle modules.
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

# Copy source code.
COPY . .

# Target command binary will be passed as a build argument when constructing creating the container.
ARG TARGET_CMD
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/app ./cmd/${TARGET_CMD}

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata docker-cli

# Create non-root user.
RUN adduser -D -g '' appuser

WORKDIR /home/appuser
COPY --chown=appuser:appuser --from=builder /bin/app .

# Run as non-root user.
USER appuser

CMD ["./app"]