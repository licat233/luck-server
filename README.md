# luck-server
抽奖小服务
## 配置：
+ 需要使用Redis緩存用戶id與ip；
+ 每個id只能抽1次；
<!-- + 每個ip最多抽5次； -->

## 业务需求:  
* 每个人只能抽一次（目前采用的line id或者填写手机号的方式并不明智，客户可以随意填写，多次尝试抽奖）

## docker安装教程：
1. docker build -t luckserver:latest .
2. docker run -d -p 81:80 --name luckserver luckserver:latest
