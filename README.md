# 自己动手写Docker实践工程
***

## 环境说明
本文基于下面的环境进行开发发运：

- Ubuntu 20 TLS ：本地搭建的Ubuntu系统
- Centos7 ：腾讯云服务器，也能跑本工程下的所有代码

注：Windows不能运行该工程，因为其中有些库是Linux采用的，

但如果想要写Windows的话，可以仓库RunC中关于Windows相关的代码

如果后期有时间的话，本工程也尝试适配下Windows系统

## 运行说明
docker demo 的代码都位于文件夹：mydocker下

可参考下面的方式运行：

```shell
go mod init dockerDemo
go mod tidy
go build mydocker/main.go
./main run -ti /bin/sh
```

## 实践文档
- [自己动手写Docker系列 -- 3.1构造实现run命令版本的容器](docs/3.1构造实现run命令版本的容器.md)
- [自己动手写Docker系列 -- 3.2增加容器资源限制](docs/3.2增加容器资源限制.md)

## 参考资料
- 《自己动手写Docker》：非常好的书籍，值得一看并实操