# PPanel 部署文档

### 服务端

运行环境

| 模块           | 需求                                 |
| :------------- | ------------------------------------ |
| 服务器配置     | 最低配置：1U2G  推荐配置: 2U4G       |
| MySQL          | MySQL 5.7及以上版本(推荐使用MySQL 8) |
| Redis          | Redis 6及以上                        |
| NGINX / Apache | 不限版本                             |

Docker环境安装文档：https://yeasy.gitbook.io/docker_practice/install

下载服务端文件:

``` shell
$ cd /root
$ wget https://github.com/perfect-panel/server/releases/download/0.1.0_alpha/ppanel.zip
$ unzip ppanel.zip
```

修改配置文件

```shell
$ cd ppanel/app
$ vim ppanel.yaml

Host: 0.0.0.0 # 监听IP
Port: 8080 # 运行端口
Debug: true # debug 模式

Logger:
  File: ./ppanel.log # 日志文件
  Level: debug # 日志等级 info debug error

JwtAuth:
  AccessSecret: 333ab3d5-429f-4a51-b001-ef484bf11db7 # Token 秘钥 (务必修改)
  AccessExpire: 604800 # Token有限期

MySQL:
  Path: 127.0.0.1 # 
  Port: 3306	
  Dbname: vpnboard # 数据库名称
  Username: root #替换为你的用户名
  Password: Password #替换为你的 root 密码
  MaxIdleConns: 10
  MaxOpenConns: 10
  LogMode: "dev"
  LogZap: false
  Config: charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai

Redis:
  Host: 127.0.0.1:6379
  Type: server
  Pass:
```



通过Docker安装环境:

docker-compose.yml

```yaml

version: '3.8'

services:
  mysql:
    image: mysql:8.0.23
    container_name: mysql_db
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword  		# 替换为你的 root 密码
      MYSQL_DATABASE: my_database          	# 默认数据库
      MYSQL_USER: user                     	# 替换为你的用户名
      MYSQL_PASSWORD: userpassword         	# 替换为你的用户密码
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7
    container_name: redis_cache
    restart: always
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  nginx:
    image: uozi/nginx-ui:latest
    container_name: nginx_server
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /root/ppanel/nginx:/etc/nginx
    depends_on:
      - mysql
      - redis

  ppanel:
    image: ppanel
    container_name: ppanel
    restart: always
    command: ["./ppanel", "run", "--config", "./ppanel.yaml"]
    volumes:
      - ./root/ppanel/app:/app
    depends_on:
      - mysql
      - redis

volumes:
  mysql_data:
  redis_data:
```

运行程序:

```shell
$ # 创建 docker-compose.yml
$ vim docker-compose.yml
$ # 粘贴上面 docker-compose.yml 内容即可
	# :wq! 保存退出
$ docker-compose up -d # 运行Docker容器
```



