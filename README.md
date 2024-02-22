# idpool
idpool 是一個 ID 產生器, 而且可以很輕易的水平擴展

## 取得 ID 的流程
![takeid_flow.drawio.png](https://github.com/Krados/idpool/blob/master/takeid_flow.drawio.png)

## 使用 docker 來建立一個本地開發環境

### 建立一個自定義網路

```
# create a user-defined bridge network
$ docker network create taipei
```

### 準備 mysql container

```
# run a mysql
$ docker run -itd --name mysql --network taipei -e MYSQL_ROOT_PASSWORD=passwd mysql

# ssh into mysql container
$ docker exec -it mysql bash

# login to mysql
$ mysql -u root -p

# create database
$ CREATE DATABASE idpool;
$ USE idpool;

# create table
$ CREATE TABLE IF NOT EXISTS `id_pool` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `start_pos` bigint unsigned NOT NULL,
  `end_pos` bigint unsigned NOT NULL,
  `current_pos` bigint unsigned NOT NULL,
  PRIMARY KEY (`id`)
);

$ CREATE TABLE IF NOT EXISTS `id_pool_claim` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `start_pos` bigint NOT NULL,
  `end_pos` bigint NOT NULL,
  `claimed_at` bigint NOT NULL,
  PRIMARY KEY (`id`)
);

$ INSERT INTO `id_pool` (`id`, `start_pos`, `end_pos`, `current_pos`) VALUES (NULL, '1', '18446744073709551615', '1');
```

### 建立 idpool container

```
# build your service
$ make docker-image

# run your service
$ docker run -itd --name idpool1 --network taipei -p 8087:8080 idpool
```

### 使用 curl 測試

```
$ curl -X POST http://localhost:8087/api/v1/newID
{"id":7,"status":200}
```

### 清理環境

````
$ docker rm -f mysql idpool1
$ docker network rm taipei
````
## 水平擴展測試

測試方式:使用 nginx 當作 load balancer, 並啟動 3 個 idpool container 當作 upstream
![scale_id_pool.png](https://github.com/Krados/idpool/blob/master/scale_id_pool.png)

### 啟動 3 個 idpool container

```
$ docker run -itd --name idpool1 --network taipei idpool
$ docker run -itd --name idpool2 --network taipei idpool
$ docker run -itd --name idpool3 --network taipei idpool
```

### 啟動 nginx container

```
# 新增一個我們需要的 default.conf
$ echo 'upstream myapp {
    server idpool1:8080;
    server idpool2:8080;
    server idpool3:8080;
}

server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;
    location / {
        proxy_pass http://myapp;
    }
    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }
}
' >> default.conf

# 啟動 nginx 並使用 default.conf
$ docker run --name nginx -v "/$PWD/default.conf":/etc/nginx/conf.d/default.conf --network taipei -d -p 8090:80 nginx
```

### 使用 curl 測試

```
$ curl --location --request POST 'localhost:8090/api/v1/newID'
{"id":21016,"status":200}
$ curl --location --request POST 'localhost:8090/api/v1/newID'
{"id":22012,"status":200}
$ curl --location --request POST 'localhost:8090/api/v1/newID'
{"id":23012,"status":200}
```

### 清理環境

```
$ docker rm -f idpool1 idpool2 idpool3 nginx
```
