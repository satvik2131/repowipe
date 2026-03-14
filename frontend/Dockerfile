# React Frontend Dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
COPY vite.config.ts ./
RUN npm install --frozen-lockfile
COPY . .
ARG BUILD_CMD=build
RUN npm run ${BUILD_CMD}

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY package*.json ./
RUN npm install --production --frozen-lockfile
EXPOSE 3000
ENV VITE_API_URL=http://backend:8080/api
CMD ["npx", "serve", "-s", "dist", "-l", "3000"]
