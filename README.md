<!--
 * @Author: flwfdd
 * @Date: 2023-03-15 15:19:46
 * @LastEditTime: 2025-01-26 15:55:11
 * @Description: _(:з」∠)_
-->
# BIT101-GO

<div align="center">

[BIT101主仓库](https://github.com/BIT101-dev/BIT101) · [API文档](https://bit101-api.apifox.cn) · [运维手册](https://bit101-project.feishu.cn/wiki/ID4owkKQKi3E0OkNV1gcfzqynHd)

</div>

`BIT101`的服务端，基于`Go`语言。

如果其他同学有需要用到该项目，可以直接使用，但需要注明使用了该项目，另外还请注意开源协议。文档中的说明已经比较详尽了，如有问题欢迎提`issue`。

🥳也非常欢迎你来贡献代码！

总之，希望这个项目能够帮助到大家( ´▽｀)


## 部署指南

一些更详细的说明可参考[运维手册](https://bit101-project.feishu.cn/wiki/ID4owkKQKi3E0OkNV1gcfzqynHd)。

### 配置依赖服务

`BIT101-GO`主要使用了`PostgreSQL`数据库、`Meilisearch`推荐系统和`Gorse`推荐系统三大依赖服务，你可以选择使用`docker`或手动安装的方式，推荐使用`docker`方式，只需要进入到`env`目录并运行`docker compose up`即可一键运行所有依赖服务，关于更多`docker`的相关操作可自行查询。

注意需要同步`docker-compose.yaml`和`config.yml`中的配置。

默认情况况下，将在几个端口上提供服务：
* `7700`：`meilisearch`的端口，除了提供`API`外，浏览器访问还有交互式搜索测试网页
* `8086`：`gorse`的`gRPC`端口
* `8088`：`gorse`的`HTTP`端口，浏览器访问显示后台网页
* `54320`：`PostgreSQL`数据库端口

### 启动主程序

需要先配置好依赖服务。

* 克隆并进入代码仓库
* 安装`GO`并配置
* 复制`config_example.yml`为`config.yml`并配置
* 执行`go run .`以启动调试
* 执行`go build -o main main.go`以编译为可执行文件`main`

假设编译后程序为`main`
* 开启服务：`./main`或`./main server`
* 备份数据：`./main bakcup`，默认备份到`./data/backup/`文件夹下
* 导入课程表：`./main import_course [path]`，其中`path`为课程表`.csv`文件所在的目录，默认为`./data/course/`
* 获取课程历史：`./main course_history [start_year] [end_year] [webvpn_cookie]`，现阶段测试最早的课程历史为`2005-2006`学年


### 数据库手动配置

使用`postgresql`作为数据库。

```sql
CREATE USER bit101 WITH PASSWORD '*****';
CREATE DATABASE bit101 OWNER bit101;
GRANT ALL PRIVILEGES ON DATABASE bit101 TO bit101;
```

### 搜索系统手动配置

使用`Meilisearch`作为搜索系统。可前往[官网链接](https://www.meilisearch.com/)查看安装和实用教程。

部署时运行：
```shell
meilisearch --env production --master-key BIT101 --db-path ./data/meilisearch
```

其中`master key`需要和`config.yml`中对应。

### 推荐系统手动部署

1. 前往[Gorse仓库Release](https://github.com/gorse-io/gorse/releases/)下载对应平台`latest`版本
2. 修改配置文件`./env/gorse_config.toml` 
3. 运行`gorse-in-one`:
```shell
./gorse-in-one -c ./env/config.toml
```

### 身份验证说明

使用`JWT`作为身份验证的方式，这样服务器就不用缓存任何信息，只要验证数字签名就可以，数字签名的密钥在`config.yml`中配置，绝对不可以泄漏。登录成功时会下发一个`fake-cookie`，之后只需要每次请求携带`fake-cookie`头即可。

### 部署到 Serverless 以及代理请求
`serverless`的分支中为专供`Serverless`平台使用的精简版本，当前只用于成绩查询。

`BIT101-GO`可以非常简单地部署到函数计算服务上，也可以按照路由将请求转发到另一个服务器上，实现反向代理。

一个典型的应用是，查询成绩详情需要进行大量的请求，这些请求如果在同一台机器上完成可能会造成比较大的压力，于是就可以将成绩查询接口的请求转发给函数计算代理完成。

如果只是快捷开关并配置代理链接，更改配置文件中的`proxy`字段即可，但如果要定义哪些接口需要走代理，需要在`router`中添加获移除特定路由的`Proxy`中间件（默认已经在部分接口添加了代理中间件）。

#### 腾讯云部署

在`Serverless`中新建函数服务，选择运行环境为`GO`的`Web函数`。

然后使用命令`GOOS=linux GOARCH=amd64 go build -o main main.go`编译为目标平台的可执行文件（也可以在相同平台上直接编译），再将生成的`main`可执行文件、`config.yaml`和一个文件名为`scf_bootstrap`的文件打成一个压缩包上传即可。其中`scf_bootstrap`文件的内容如下：
```bash
#!/bin/bash

./main
```
这样就将`BIT101-GO`部署到了`Serverless`平台上。

另外需要注意函数配置中几个参数的设置：超时时间、请求多并发等。

### 生产环境部署最佳实践

当前部署在`Ubuntu`上，使用`screen`来实现后台运行。

首先按照之前的说明跑起依赖服务并编译好`main`程序，然后创建`production`文件夹，将`config_example.yml`复制为`production/config.yml`并进行配置，同时将`main`也复制到`production`文件夹下。

使用`screen -S bit101-go`创建一个后台运行控制台，并使用`screen -r bit101-go`进入，然后`cd`到`production`文件夹下，使用`bash run.sh`运行程序。


## 性能测试

使用`ab -n 1000 -c 100`进行压力测试，为排除网络影响，均为同一台服务器本机运行测试。

首先测试了一个比较复杂的接口。

老`Python`后端结果：

```shell
Document Path:          /reaction/comments/?obj=paper1&order=default&page=0
Document Length:        17314 bytes

Concurrency Level:      100
Time taken for tests:   44.034 seconds
Complete requests:      1000
Failed requests:        0
Total transferred:      17494000 bytes
HTML transferred:       17314000 bytes
Requests per second:    22.71 [#/sec] (mean)
Time per request:       4403.378 [ms] (mean)
Time per request:       44.034 [ms] (mean, across all concurrent requests)
Transfer rate:          387.97 [Kbytes/sec] received
```

新`GO`后端结果：
```shell
Document Path:          /reaction/comments?obj=paper1&order=default&page=0
Document Length:        18019 bytes

Concurrency Level:      100
Time taken for tests:   2.939 seconds
Complete requests:      1000
Failed requests:        106
   (Connect: 0, Receive: 0, Length: 106, Exceptions: 0)
Total transferred:      18121202 bytes
HTML transferred:       18018202 bytes
Requests per second:    340.31 [#/sec] (mean)
Time per request:       293.851 [ms] (mean)
Time per request:       2.939 [ms] (mean, across all concurrent requests)
Transfer rate:          6022.28 [Kbytes/sec] received
```

一通测试猛如虎，结果非常 AMAZING 啊，老、新后端在该接口上的秒并发分别为`22.71`、`340.31`，提升了约**15倍**！！！（然而当前实际场景中服务器带宽反而成为瓶颈了）

又测试了一个比较简单的接口（仅包含一个数据库查询），老、新后端秒并发分别为`708`、`955`，仍然一定的提升。
