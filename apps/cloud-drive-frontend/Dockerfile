FROM node:22-alpine AS builder

WORKDIR /workspace

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY pnpm-lock.yaml pnpm-workspace.yaml package.json ./
COPY apps/cloud-drive-frontend/package.json ./apps/cloud-drive-frontend/package.json

RUN pnpm install --frozen-lockfile --filter cloud-drive-frontend...

COPY apps/cloud-drive-frontend ./apps/cloud-drive-frontend

RUN pnpm --filter cloud-drive-frontend build

FROM nginx:1.29-alpine

COPY apps/cloud-drive-frontend/nginx.conf /etc/nginx/templates/default.conf.template
COPY --from=builder /workspace/apps/cloud-drive-frontend/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
