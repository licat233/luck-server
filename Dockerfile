FROM golang:1.17.3-alpine3.15
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN apk add redis;apk add supervisor;go env -w GO111MODULE=on;go env -w GOPROXY=https://goproxy.cn,direct
WORKDIR /app
COPY . .
COPY supervisord.conf /etc/supervisord.conf
RUN go mod tidy && CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o goluck .
CMD ["supervisord"]