# 第一阶段：构建阶段
FROM golang:alpine AS builder

# 标记这是构建阶段
LABEL stage=gobuilder

# 设置Go编译环境
ENV CGO_ENABLED 0
# 设置Go模块代理，加速依赖下载
ENV GOPROXY https://goproxy.cn,direct

# 使用清华镜像源加速Alpine包管理
RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories

# 更新Alpine包并安装时区数据
RUN apk update --no-cache && apk add --no-cache tzdata

# 设置工作目录
WORKDIR /build

# 复制Go项目依赖文件
ADD go.mod .
ADD go.sum .

# 安装git用于下载依赖
RUN apk add --no-cache git

# 下载Go依赖
RUN go mod download

# 复制项目源代码和配置文件
COPY . .
COPY ./etc /app/etc

# 编译Go项目，使用ldflags减小二进制体积
RUN go build -ldflags="-s -w" -o /app/grpc_client .

# 第二阶段：运行阶段，使用scratch作为基础镜像以减小体积
FROM scratch

# 复制SSL证书，用于HTTPS请求
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# 复制时区信息
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
# 设置时区环境变量
ENV TZ Asia/Shanghai

# 设置工作目录
WORKDIR /app
# 从构建阶段复制编译好的二进制文件和配置
COPY --from=builder /app/grpc_client /app/grpc_client
COPY --from=builder /app/etc /app/etc

# 设置容器启动命令
CMD ["./grpc_client", "-f", "etc/client.yaml"]
