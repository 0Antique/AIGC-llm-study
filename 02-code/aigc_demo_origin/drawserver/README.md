# DEMO服务端
包含AI画画服务端和小程序服务端

## 编译
```shell
cd drawer
go build .
```

## 启动服务：
启动小程序服务端和AI画画服务： 
```shell
./drawserver -w=true
```

仅启动小程序服务端：
```shell
./drawserver 
```



在终端执行以下内容：

1、切换当前工作目录 "drawer"中

```
cd drawer
```

2、在进入 "drawer" 目录之后，执行以下命令：

```
go build .
```

 Go 编译器在当前目录中构建 Go 程序。点号（`.`）表示当前目录。`go build` 命令会编译当前目录中的 Go 源代码文件，并生成一个可执行的二进制文件。

3、运行 "drawserver" 的可执行文件，启动小程序服务端和AI画画服务

```
./drawserver -w=true
```

