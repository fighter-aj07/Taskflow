# Stage 1: Build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
# Remove local vendor_local replace directives (vendor/ has all deps)
# In modules.txt: strip " => path" from versioned lines, delete bare replace lines
RUN sed -i '/^replace .*/d' go.mod && \
    sed -i '/^# [a-z].* v[0-9].* => /s/ => .*//' vendor/modules.txt && \
    sed -i '/^# [a-z].* => /d' vendor/modules.txt
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /api ./cmd/api

# Stage 2: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /api /api
COPY migrations/ /migrations/
COPY seeds/ /seeds/
EXPOSE 8080
CMD ["/api"]
