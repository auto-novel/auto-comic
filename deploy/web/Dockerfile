FROM node:22-alpine AS deps
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN corepack install
COPY apps/admin/package.json ./apps/admin/package.json
COPY apps/web/package.json ./apps/web/package.json
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile

FROM deps AS builder
WORKDIR /app
ARG VITE_COMMIT_SHA
ARG VITE_BUILD_TIME
COPY apps/admin/. ./apps/admin/
COPY apps/web/. ./apps/web/
RUN VITE_COMMIT_SHA="${VITE_COMMIT_SHA}" \
    VITE_BUILD_TIME="${VITE_BUILD_TIME}" \
    pnpm run build

FROM caddy:2.7.4
COPY --from=builder /app/apps/admin/dist /admin
COPY --from=builder /app/apps/web/dist /web
COPY deploy/web/Caddyfile /etc/caddy/Caddyfile
