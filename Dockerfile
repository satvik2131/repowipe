# Stage 1 — Build frontend
FROM node:latest AS frontend-build
WORKDIR /frontend
COPY frontend/package*.json ./
COPY frontend/vite.config.ts ./
RUN npm install
COPY frontend ./
RUN npm run build

# Stage 2 — Build Go backend
FROM golang:latest AS backend-build
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download 

COPY backend ./
COPY --from=frontend-build /frontend/dist ./static
RUN go mod tidy
RUN go build -o repowipe

# Stage 3 — Final small image
FROM alpine:latest
WORKDIR /app

# Copy the binary
COPY --from=backend-build /backend/repowipe .

# Copy static files (THIS WAS MISSING!)
COPY --from=backend-build /backend/static ./static

EXPOSE 8080
CMD ["./repowipe"]