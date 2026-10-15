# syntax=docker/dockerfile:1

# =============================================================
# mini-kanban-web（Next.js 14 App Router）
# 构建：docker build -t mini-kanban-web ./web
#
# 注意：项目未开启 output: "export"，next build 产出的是 Node 运行时
# 产物（.next），因此运行阶段必须用 next start 常驻，不能只丢静态文件。
# =============================================================

# ---------- 依赖（含 devDependencies：tailwind / typescript 参与构建） ----------
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

# ---------- 构建 ----------
FROM node:20-alpine AS builder
WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY . .

# NEXT_PUBLIC_* 是构建期内联的：改了地址必须重新构建镜像，运行时再设无效
ARG NEXT_PUBLIC_API_BASE=/api
ENV NEXT_PUBLIC_API_BASE=${NEXT_PUBLIC_API_BASE} \
    NEXT_TELEMETRY_DISABLED=1

RUN npm run build

# ---------- 运行期依赖（只装生产依赖，去掉编译期工具链） ----------
FROM node:20-alpine AS prod-deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --omit=dev

# ---------- 运行 ----------
FROM node:20-alpine AS runner

RUN apk add --no-cache tzdata wget \
 && addgroup -g 10001 -S nextjs \
 && adduser  -u 10001 -S -G nextjs -H nextjs

ENV NODE_ENV=production \
    TZ=Asia/Shanghai \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3201

WORKDIR /app

COPY --from=prod-deps /app/node_modules   ./node_modules
COPY --from=builder   /app/.next          ./.next
COPY --from=builder   /app/package.json   ./package.json
COPY --from=builder   /app/next.config.mjs ./next.config.mjs

# next start 运行期会往 .next/cache 写缓存，而上面的 COPY 以 root 落地，
# 这里把必要目录交给运行用户，避免启动后报权限错误
# （只递归 cache 子目录，node_modules 保持原属主以免镜像层膨胀）
RUN mkdir -p /app/.next/cache \
 && chown nextjs:nextjs /app /app/.next \
 && chown -R nextjs:nextjs /app/.next/cache

USER nextjs

# 只在内网被 nginx 访问，不映射到宿主机
EXPOSE 3201

HEALTHCHECK --interval=15s --timeout=3s --start-period=20s --retries=5 \
  CMD wget -q -O /dev/null http://127.0.0.1:3201/login || exit 1

# 用 exec 形式直接启动 next：npm 作为 PID 1 会吞掉 SIGTERM，导致停容器只能强杀
CMD ["node", "node_modules/next/dist/bin/next", "start", "-p", "3201"]
