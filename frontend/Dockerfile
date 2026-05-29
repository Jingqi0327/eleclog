FROM nginx:alpine

# 将自定义的 nginx 配置文件复制到容器中
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 将前端所有的静态文件复制到 Nginx 的默认提供服务的目录下
COPY html/ /usr/share/nginx/html/

# 暴露 80 端口
EXPOSE 80

# 启动 Nginx
CMD ["nginx", "-g", "daemon off;"]
