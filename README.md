# idpool
idpool 是一個 ID 產生器, 而且可以很輕易的水平擴展

## 何時會使用到此專案
大部分情境下可以直接使用 [snowflake](https://github.com/bwmarrin/snowflake) 或者是 [sonyflake](https://github.com/sony/sonyflake) 來做全局 ID 產生器, 但者兩者皆是以時間為驅動, 就代表著有絕大多數的 ID 不會被使用到.

如果想要有效的使用到所有 ID 時可以嘗試使用此專案.

## 實作想法

server 為冷啟動, 即每次啟動時並不會直接從 id pool 中取得 1000 筆資料, 而是當 server 接收到 take id 的請求才取從 id pool 取得 1000 筆資料並寫入到 queue 內.

如果 queue 裡面沒有可用 id, 就會再一次去 id pool 中取得 1000 筆資料並寫入到 queue 內.

讀取和寫入的比例為 1:999

queue 是使用 golang 的 channel 來實現.

使用 mysql 當作紀錄可用 id pool 紀錄的地方.

不需要使用到 cache, 因為每次請求皆會返回不同的 ID.

server 為 stateless 所以水平擴展方便.

## 可能的問題

server 重啟時即便 queue 裡面還有可用的 id 皆會直接丟棄, 最慘的情況會直接丟棄 999 筆 id

mysql 為 server 取得 id 的地方, 以目前專案的實作方式會有單點故障(single point of failure)的問題.

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
