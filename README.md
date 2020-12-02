# debug_tool

这是一个辅助php调试输出的工具，需要配合 php composer 库 `wj008/debug` 使用。

`composer require wj008/debug`

#### 1.设计思路

由于php调试输出的时候，在很多情况下，上线以后调试非常麻烦。

通常php开发者会这样做

```php
<?php
.....
function func1($data){
......
    foreach ($data as $item){
      print_r($item)
      exit();
      ......
    }
......
}
```
##### 这样的弊端

 1. 首先会使得流程被中断，for循环中的其他项无法得到执行。
 2. 会干扰页面输出，或者是json 输出导致前端报错，前端操作无法继续。
 3. 会在上线的时候可能存在遗忘注释掉调试代码，而影响线上代码质量。
 4. 服务器上线后无法远程调试输出。
 
##### 如何解决这个问题

1. 我们通过  php 的Logger类来替代 print_r() 等输出。
php 代码仓库地址 `https://github.com/wj008/debug` composer 安装命令 `composer require wj008/debug`

```php
<?php
use debug\Logger;
.....
function func1($data){
......
    foreach ($data as $item){
     Logger::log($item);
      ......
    }
......
}
```

使用 `Logger::log`,`Logger::info`,`Logger::error` 等方法来输出信息。

Logger 会将使用 php socket UDP 将调试`$item`数据整理后发送到  go 开发的  `debug_tool` 中,在 `debug_tool` 中美化后在命令行中打印出来。

整个流程不阻塞，也不会中断php代码执行，不对页面输出任何东西，不干扰输出结果。

#### 2 debug_tool 的使用

配置文件：
`debug.env`
```
#主服务端口地址信息
server_tcp=0.0.0.0:8000
#用于接收PHP 的udp 调试信息的端口地址
server_udp=0.0.0.0:1024

#双方通讯秘钥，防止其他人连入服务，尽可能的设置复杂一些，
password=sdskljhxdiuye8u2i3o928y378o92309ohjbsdgf9o2309@8623
#服务器端如果配置了日志文件路径，则会生成相应的日志文件保存。按天生成。
log_info_file=logs/info_{date}.log
log_error_file=logs/error_{date}.log

#作为客户端时 用于连接服务器的ip地址端口
client_tcp=127.0.0.1:8000

```

`debug_tool` 是一个可以用于保存打印日志的服务运行，也可以作为连接服务的客户端运行。

例如 我们需要在已经上线的项目中使用远程调试输出信息。
我们可以在web服务器（S服务器）上运行，php 的日志将会发送到(S服务器)调试服务器

作为服务端
`debug.env` 
```
#主服务端口地址信息
server_tcp=0.0.0.0:8000
#用于接收PHP 的udp 调试信息的端口地址
server_udp=0.0.0.0:1024

#双方通讯秘钥，防止其他人连入服务，尽可能的设置复杂一些，
password=sdskljhxdiuye8u2i3o928y378o92309ohjbsdgf9o2309@8623
#服务器端如果配置了日志文件路径，则会生成相应的日志文件保存。按天生成。
log_info_file=logs/info_{date}.log
log_error_file=logs/error_{date}.log

```

```$ ./debug_tool -s```

然后我们的开发本地机器 想获得 （S服务器）在运行时的调试数据，并希望实时打印，我们只需要配置好  `debug.env` 然后运行
作为客户端
`debug.env` 
```
#主服务端口地址信息(这个需要保留)
server_tcp=0.0.0.0:8000
#用于接收PHP 的udp 调试信息的端口地址
server_udp=0.0.0.0:1024

#双方通讯秘钥，与服务端保持一致方可通讯
password=sdskljhxdiuye8u2i3o928y378o92309ohjbsdgf9o2309@8623
#作为客户端时 用于连接服务器的ip地址端口
client_tcp=127.0.0.1:8000
```
```$ ./debug_tool -c```


![Image text](https://raw.githubusercontent.com/wj008/debug_tools/main/img/QQ20201202-1.png)


当然 也可以直接在本地调试：

作为客户端
`debug.env` 
```
#主服务端口地址信息(这个需要保留)
server_tcp=0.0.0.0:8000
#用于接收PHP 的udp 调试信息的端口地址
server_udp=0.0.0.0:1024

#双方通讯秘钥，与服务端保持一致方可通讯
password=sdskljhxdiuye8u2i3o928y378o92309ohjbsdgf9o2309@8623
#作为客户端时 用于连接服务器的ip地址端口
client_tcp=127.0.0.1:8000
```
命令行不带参数即可

```$ ./debug_tool```

![Image text](https://raw.githubusercontent.com/wj008/debug_tools/main/img/QQ20201202-0.png)
