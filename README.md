# luck-server
抽奖系统服务端
## 配置：
+ 需要使用Redis緩存用戶id與ip；
+ 每個id只能抽1次；
<!-- + 每個ip最多抽5次； -->

## docker安装教程：
1. docker build -t luckserver:latest .
2. docker run -d -p 81:12345 --name luckserver luckserver:latest