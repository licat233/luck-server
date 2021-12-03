FROM golang:1.17.3-alpine3.15
WORKDIR /app
COPY . .
COPY supervisord.conf /etc/supervisord.conf
EXPOSE 80
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN apk add redis;apk add supervisor;go env -w GO111MODULE=on;go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod tidy && go build -o luckserver
CMD ["supervisord","-c","/etc/supervisord.conf"]