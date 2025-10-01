# Stage 1: Build the Svelte application
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package.json. A package-lock.json is highly recommended for reproducible builds.
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the source code
COPY . .

# Build the application for production
RUN npm run build

# Stage 2: Serve the application with Nginx
FROM nginx:1.25-alpine

# Copy the built static files from the builder stage to the Nginx html directory
COPY --from=builder /app/build /usr/share/nginx/html

# Copy the custom Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Nginx exposes port 80 by default, which is what Traefik is configured to connect to.
# No EXPOSE needed here as it's documentation.

# Start Nginx in the foreground
CMD ["nginx", "-g", "daemon off;"]