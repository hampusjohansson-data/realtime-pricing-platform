# --- Build stage ---
FROM node:22-alpine AS build

WORKDIR /app

# Installera dependencies
COPY package*.json ./
RUN npm ci

# Kopiera resten av frontenden
COPY . .

# Bygg Vite (prod-build)
RUN npm run build

# --- Runtime stage (nginx) ---
FROM nginx:1.27-alpine

# Ta bort default-konfig
RUN rm /etc/nginx/conf.d/default.conf

# Vår egen nginx-konfig (se nginx.conf nedan)
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Kopiera byggda filer
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
