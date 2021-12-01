FROM golang:1.17.3-alpine3.15
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
ENV REDIS_HOME /usr/local
RUN apk add redis && go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct
RUN /usr/bin/redis-server /etc/redis.conf &
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o goluck .
# ENTRYPOINT ["./goluck"]