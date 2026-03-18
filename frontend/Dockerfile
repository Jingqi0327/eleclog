# --- 第一阶段：编译环境 (Builder) ---
FROM node:24.14-alpine3.22 AS builder
WORKDIR /app
# 先复制 package.json 依靠 Docker 缓存加速安装
COPY package*.json ./
RUN npm install
# 复制源码并执行打包
COPY . .
RUN npm run build


# --- 第二阶段：运行环境 (Nginx) ---
FROM nginx:1.29-alpine
# 删除 Nginx 默认配置
RUN rm /etc/nginx/conf.d/default.conf
# 复制自定义配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 从第一阶段 (builder) 复制编译好的 dist 到 Nginx 目录
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]