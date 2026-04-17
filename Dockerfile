# 第一阶段：编译阶段
#FROM golang:1.21-alpine AS builder
FROM registry.cn-hangzhou.aliyuncs.com/acs/golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制所有源码和资源
COPY . .

# 初始化模块并编译 (由于是克隆的源码，确保这里能跑通)
RUN go mod init viewghost || true
RUN go mod tidy
# 编译生成名为 viewghost 的 Linux 二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -o viewghost .

# 第二阶段：运行阶段
#FROM alpine:latest
FROM registry.cn-hangzhou.aliyuncs.com/acs/alpine:latest

WORKDIR /app

# 从编译阶段拷贝二进制文件
# 将二进制文件拷贝到系统路径，这样无论 /app 怎么挂载，程序都能找到
COPY --from=builder /app/viewghost /usr/local/bin/viewghost

# 拷贝示例配置文件（或者由用户挂载）
#COPY --from=builder /app/viewghost.config.example ./viewghost.config

# 暴露端口 (根据你的默认配置，通常是 8080)
EXPOSE 8080

# 运行程序
ENTRYPOINT ["viewghost"]