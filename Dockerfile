# syntax=docker/dockerfile:1
# ---- Base ----
FROM node:20-alpine AS base
WORKDIR /app
RUN apk add --no-cache python3 make g++ openssl

# ---- Dependencies ----
FROM base AS deps
COPY package.json package-lock.json* ./
RUN npm install

# ---- Build ----
FROM deps AS build
COPY tsconfig.json ./
COPY prisma ./prisma
COPY src ./src
RUN npx prisma generate
RUN npm run build

# ---- Production runtime ----
FROM node:20-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
RUN apk add --no-cache openssl
COPY package.json package-lock.json* ./
COPY --from=deps /app/node_modules ./node_modules
COPY --from=build /app/dist ./dist
COPY --from=build /app/prisma ./prisma
EXPOSE 4000
# Default command runs the API. Override to `npm run start:workers` for workers.
CMD ["node", "dist/server.js"]
