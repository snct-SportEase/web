# Stage 1: Build the SvelteKit application
FROM node:20-alpine AS builder

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm install

# Copy source code and build
COPY . .
RUN npm run build

# Stage 2: E2E test stage
FROM builder AS e2e

RUN apk add --no-cache chromium nss freetype harfbuzz ca-certificates ttf-freefont

ENV PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH=/usr/bin/chromium-browser
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1

CMD ["npm", "run", "test:e2e"]

# Stage 3: Production stage
FROM node:20-alpine AS production

WORKDIR /app

# Copy built assets from builder stage
COPY --from=builder /app/build ./build
RUN ln -s ../../client ./build/server/chunks/client
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./

# Expose the port the app will run on
EXPOSE 3300

# Set the command to start the server
# The server will listen on 0.0.0.0 by default through HOST=0.0.0.0
ENV HOST=0.0.0.0
CMD ["node", "build/index.js"]
