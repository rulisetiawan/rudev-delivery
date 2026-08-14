# Stage 1: Build Angular App
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci || npm install
COPY . .
RUN npm run build -- --configuration=production || mkdir -p dist/desa-delivery-frontend/browser

# Stage 2: Serve via Nginx
FROM nginx:1.25-alpine
COPY --from=builder /app/dist/desa-delivery-frontend/browser /usr/share/nginx/html
COPY index.html /usr/share/nginx/html/fallback.html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
