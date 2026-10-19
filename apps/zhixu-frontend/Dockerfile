FROM node:22-alpine AS builder

WORKDIR /workspace

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY pnpm-lock.yaml pnpm-workspace.yaml package.json ./
COPY apps/zhixu-frontend/package.json ./apps/zhixu-frontend/package.json

RUN pnpm install --frozen-lockfile --filter zhixu-frontend...

COPY apps/zhixu-frontend ./apps/zhixu-frontend

RUN pnpm --filter zhixu-frontend build

FROM nginx:1.29-alpine

COPY apps/zhixu-frontend/nginx.conf /etc/nginx/templates/default.conf.template
COPY --from=builder /workspace/apps/zhixu-frontend/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
