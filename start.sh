#!/bin/sh
# alpine 镜像默认没有 bash，使用 sh 来编写启动脚本

# 设置脚本在遇到错误时立即退出，确保任何命令失败都会导致整个脚本停止执行
set -e

echo "start app..."
# 执行传递给脚本的命令，通常是启动应用程序的命令
# 也就是CMD ["/app/main"]里的命令
exec "$@"