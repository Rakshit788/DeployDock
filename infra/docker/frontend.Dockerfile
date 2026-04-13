# Stage 1: Build
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY apps/frontend/package.json apps/frontend/package-lock.json* ./

# Install dependencies
RUN npm install

# Copy source
COPY apps/frontend/ .

# Build Next.js app
RUN npm run build

# Stage 2: Runtime
FROM node:20-alpine

WORKDIR /app

# Install only production dependencies
COPY apps/frontend/package.json apps/frontend/package-lock.json* ./
RUN npm install --production

# Copy built app from builder
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/next.config.js ./

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD node -e "require('http').get('http://localhost:3000', (r) => {if (r.statusCode !== 200) throw new Error(r.statusCode)})"

# Start
CMD ["npm", "start"]
