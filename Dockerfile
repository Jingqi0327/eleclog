FROM golang:1.25.8-alpine3.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

FROM alpine:3.22
WORKDIR /app
# 从 builder 阶段复制构建好的可执行文件到当前阶段
COPY --from=builder /app/main .
COPY app.env .
# 复制数据库迁移文件到当前阶段，确保应用在运行时可以访问这些文件
COPY db/migration ./db/migration
# 复制一个启动脚本到当前阶段，这个脚本会在容器启动时执行，负责先执行数据库迁移，然后再启动应用
COPY start.sh .
COPY wait-for.sh .
EXPOSE 6060
CMD ["/app/main"]
# 使用 ENTRYPOINT 来指定容器启动时执行的命令，这里我们可以使用一个 shell 脚本来先执行数据库迁移，然后再启动应用
ENTRYPOINT [ "/app/start.sh" ]