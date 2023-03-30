<!--
 * @Author: flwfdd
 * @Date: 2023-03-15 15:19:46
 * @LastEditTime: 2023-03-30 18:01:08
 * @Description: _(:з」∠)_
-->
# BIT101-GO

BIT101的新后端，基于GO。

项目仍在开发中，敬请期待。

## 部署指南

以`Ubuntu`为例。

* 安装`PostgreSQL`并配置
* 安装`GO`
* 执行`go build -o main main.go`以编译
* 配置`config.yaml`

### 数据库配置

使用`postgresql`作为数据库。

```sql
CREATE USER bit101 WITH PASSWORD '*****';
CREATE DATABASE bit101 OWNER bit101;
GRANT ALL PRIVILEGES ON DATABASE bit101 TO bit101;
```

### 部署到Serverless以及代理请求

BIT101-GO可以非常简单地部署到函数计算服务上，也可以按照路由将请求转发到另一个服务器上，实现反向代理。

一个典型的应用是，查询成绩详情需要进行大量的请求，这些请求如果在同一台机器上完成可能会造成比较大的压力，于是就可以将成绩查询接口的请求转发给函数计算代理完成。

如果只是快捷开关并配置代理链接，更改配置文件中的`proxy`字段即可，但如果要定义哪些接口需要走代理，需要在`router`中添加获移除特定路由的`Proxy`中间件（默认已经在部分接口添加了代理中间件）。

#### 腾讯云部署

在Serverless中新建函数服务，选择运行环境为`GO`的`Web函数`。

然后使用命令`GOOS=linux GOARCH=amd64 go build -o main main.go`编译为目标平台的可执行文件，再将生成的`main`和一个文件名为`scf_bootstrap`的文件打成一个压缩包上传即可。其中`scf_bootstrap`文件的内容如下：
```bash
#!/bin/bash

./main
```
这样就将BIT101-GO部署到了Serverless平台上。

另外需要注意函数配置中几个参数的设置：超时时间、请求多并发等。

## 身份验证说明

使用`JWT`作为身份验证的方式，这样服务器就不用缓存任何信息，只要验证数字签名就可以，数字签名的密钥在`config.yml`中配置，绝对不可以泄漏。登录成功时会下发一个`fake-cookie`，之后只需要每次请求携带`fake-cookie`头即可。

## 全文搜索功能实现

对于文章等内容有全文搜索的需求，但是如果直接使用数据库`LIKE %key%`查询的话效率实在太低，于是考虑使用全文索引。

实现方式依赖`PostgreSQL`数据库，首先建立一个`tsvector`类型的字段，并且在这个字段上建立`gin`索引，然后将待搜索的文本进行分词，再将分词后的结果加入到这个字段中，之后搜索时使用`tsquery`即可。
