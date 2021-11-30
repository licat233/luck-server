FROM golang:latest
WORKDIR $GOPATH/src/gin-blog
COPY . $GOPATH/src/gin-blog
ENV GOPROXY https://goproxy.cn
EXPOSE 8000
ENTRYPOINT ["go","mod","tidy"]
CMD go get github.com/mattn/go-isatty@v0.0.12
ENTRYPOINT ["go","run","server.go"]