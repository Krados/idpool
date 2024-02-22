# idpool
idpool 是一個 ID 產生器, 而且可以很輕易的水平擴展

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