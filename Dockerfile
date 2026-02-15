# Stage 1: Build frontend assets
FROM node:22-alpine AS frontend-builder

WORKDIR /build

# Build guard frontend
COPY ui/guard/package.json ui/guard/package-lock.json ./ui/guard/
RUN cd ui/guard && npm ci

COPY ui/guard/ ./ui/guard/

ARG BACKEND_HOST=http://localhost:9999/wall
ARG ROOT_PATH=/wall
ARG ADMIN_PATH=/admin
ARG GUARD_PATH=/guard

ENV VITE_BACKEND_HOST=$BACKEND_HOST \
    VITE_ROOT_PATH=$ROOT_PATH \
    VITE_ADMIN_PATH=$ADMIN_PATH \
    VITE_GUARD_PATH=$GUARD_PATH

RUN cd ui/guard && npm run build

# Build admin frontend
COPY ui/admin/package.json ui/admin/package-lock.json ./ui/admin/
RUN cd ui/admin && npm ci

COPY ui/admin/ ./ui/admin/

RUN cd ui/admin && npm run build

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS backend-builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Copy frontend dist from stage 1
COPY --from=frontend-builder /build/ui/guard/dist ./ui/guard/dist
COPY --from=frontend-builder /build/ui/admin/dist ./ui/admin/dist

ARG BACKEND_HOST=http://localhost:9999/wall
ARG ROOT_PATH=/wall
ARG ADMIN_PATH=/admin
ARG GUARD_PATH=/guard

RUN CGO_ENABLED=1 go build -ldflags "\
    -X 'main.BackendHost=$BACKEND_HOST' \
    -X 'main.RootPath=$ROOT_PATH' \
    -X 'main.AdminPath=$ADMIN_PATH' \
    -X 'main.GuardPath=$GUARD_PATH' \
    -s -w" \
    -o app main.go

# Stage 3: Minimal runtime image
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=backend-builder /build/app .
COPY config/config.yaml ./config/config.yaml

RUN mkdir -p /app/data /app/logs

ARG PORT=9999
ENV PORT=$PORT
EXPOSE $PORT

ENTRYPOINT ["./app"]
