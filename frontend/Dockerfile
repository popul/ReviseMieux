# Frontend Dockerfile
# For development: serves static files or runs dev server
# For production: builds and serves with nginx

# ============================================================================
# Development stage
# ============================================================================
FROM node:20-alpine AS development

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy source code
COPY . .

# Expose dev server port
EXPOSE 3000

# Default command for development (can be overridden)
CMD ["npm", "run", "dev"]

# ============================================================================
# Build stage
# ============================================================================
FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# ============================================================================
# Production stage
# ============================================================================
FROM nginx:alpine AS production

# Copy nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Copy built files
COPY --from=builder /app/dist /usr/share/nginx/html

# For static HTML files (current state), copy them directly
COPY *.html /usr/share/nginx/html/

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
