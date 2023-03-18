<!--
 * @Author: flwfdd
 * @Date: 2023-03-15 15:19:46
 * @LastEditTime: 2023-03-16 12:54:41
 * @Description: _(:з」∠)_
-->
# BIT101-GO

BIT101的新后端，基于GO。

项目仍在开发中，敬请期待。

## 部署到Serverless以及代理请求

BIT101-GO可以非常简单地部署到函数计算服务上，也可以按照路由将请求转发到另一个服务器上，实现反向代理。

一个典型的应用是，查询成绩详情需要进行大量的请求，这些请求如果在同一台机器上完成可能会造成比较大的压力，于是就可以将成绩查询接口的请求转发给函数计算代理完成。

### 腾讯云部署

在Serverless中新建函数服务，选择运行环境为`GO`的`Web函数`。

然后使用命令`GOOS=linux GOARCH=amd64 go build -o main main.go`编译为目标平台的可执行文件，再将生成的`main`和一个文件名为`scf_bootstrap`的文件打成一个压缩包上传即可。其中`scf_bootstrap`文件的内容如下：
```bash
#!/bin/bash

./main
```
这样就将BIT101-GO部署到了Serverless平台上。

另外需要注意函数配置中几个参数的设置：超时时间、请求多并发。

## 缓存字段备忘
 
* 注册验证码：`verify{sid}:{code}`