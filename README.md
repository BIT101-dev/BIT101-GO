<!--
 * @Author: flwfdd
 * @Date: 2023-03-15 15:19:46
 * @LastEditTime: 2023-09-23 22:16:52
 * @Description: _(:з」∠)_
-->
# BIT101-GO

`BIT101`的新后端，基于`GO`。

[API文档](https://bit101-api.apifox.cn)

如果其他同学有需要用到该项目，可以直接使用，但需要注明使用了该项目，另外还请注意开源协议。文档中的说明已经比较详尽了，如有问题欢迎提`issue`。

🥳也非常欢迎你来贡献代码！

总之，希望这个项目能够帮助到大家( ´▽｀)


## 部署指南

以`Ubuntu`为例。

* 安装`PostgreSQL`并配置
* 安装`GO`并配置
* 执行`go build -o main main.go`以编译
* 复制`config_example.yaml`为`config.yaml`并配置

### 数据库配置

使用`postgresql`作为数据库。

```sql
CREATE USER bit101 WITH PASSWORD '*****';
CREATE DATABASE bit101 OWNER bit101;
GRANT ALL PRIVILEGES ON DATABASE bit101 TO bit101;
```

### 身份验证说明

使用`JWT`作为身份验证的方式，这样服务器就不用缓存任何信息，只要验证数字签名就可以，数字签名的密钥在`config.yml`中配置，绝对不可以泄漏。登录成功时会下发一个`fake-cookie`，之后只需要每次请求携带`fake-cookie`头即可。

### 部署到 Serverless 以及代理请求

⚠️由于用于分词的`gojieba`库调用了原生`C++`，编译部署会有各种各样的奇怪错误，故该功能暂不可用，之后会考虑单独开一个专供`Serverless`的分支。

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

## 使用方法

假设编译后程序为`main`
* 开启服务：`./main`或`./main server`
* 导入课程表：`./main import_course [path]`，其中`path`为课程表`.csv`文件所在的目录，默认为`./data/course/`
* 获取课程历史：`./main course_history [start_year] [end_year] [webvpn_cookie]`，现阶段测试最早的课程历史为`2005-2006`学年

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


## 全文搜索功能实现

对于文章等内容有全文搜索的需求，但是如果直接使用数据库`LIKE %key%`查询的话效率实在太低，于是考虑使用全文索引。

实现方式依赖`PostgreSQL`数据库，首先建立一个`tsvector`类型的字段，并且在这个字段上建立`gin`索引，然后将待搜索的文本进行分词，再将分词后的结果加入到这个字段中，之后搜索时使用`tsquery`即可。

然而就结果来说，现在还不能很好的平衡相关性和一些其他参数（如点赞数）的排序，因为分词的原因，也可能会搜出很多不相关结果来。
