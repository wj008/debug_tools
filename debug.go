package main

import (
	"flag"
	"github.com/zh-five/xdaemon"
	"log"
	"net"
	"runtime"
	"runtime/debug"
	"time"
	"yee/config"
	"yee/conn"
	"yee/echo"
)

var ClientMap map[string]*conn.TcpConn
var runType int

func main() {
	//设置启动选项
	flag.Bool("v", false, " -v 或者无参数是显示执行服务")
	sv := flag.Bool("s", false, "是否以后台服务启动")
	cv := flag.Bool("c", false, "是否以客户端启动")
	flag.Parse()
	log.SetFlags(log.LstdFlags)
	isMain := config.Int("XW_DAEMON_IDX", 0)
	if isMain == 0 && *cv == false {
		//检查端口是否被占用
		serverTcp := config.String("server_tcp", "0.0.0.0:8000")
		sv, err := net.Listen("tcp", serverTcp)
		if err != nil {
			log.Println(err.Error())
			return
		}
		sv.Close()
	}
	if *sv {
		runType = 3
		MaxCount := config.Int("server_max_count", 0)
		MaxError := config.Int("server_max_error", 10000)
		//创建一个Daemon对象
		d := xdaemon.NewDaemon("server.log")
		//调整一些运行参数(可选)
		d.MaxCount = MaxCount //最大重启次数
		d.MaxError = MaxError
		//执行守护进程模式
		d.Run()
	} else if *cv {
		runType = 2
		Client()
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	runType = 1
	Serv()
}

func Serv() {
	ClientMap = make(map[string]*conn.TcpConn)
	serverTcp := config.String("server_tcp", "0.0.0.0:8000")
	serverUdp := config.String("server_udp", "0.0.0.0:1024")
	log.Print("\n" +
		"  ____\n" +
		" |  _ \\\n" +
		" | |_) |   ___    __ _    ___    ___    _ __\n" +
		" |  _ <   / _ \\  / _` |  / __|  / _ \\  | '_ \\\n" +
		" | |_) | |  __/ | (_| | | (__  | (_) | | | | |\n" +
		" |____/   \\___|  \\__,_|  \\___|  \\___/  |_| |_|\n" +
		"=====================debug====================\n")

	go TcpServer(serverTcp)
	log.Print("TCP:" + serverTcp)
	go UdpServer(serverUdp)
	log.Print("UDP:" + serverUdp)
	select {}
}

func Client() {
	log.Print("\n" +
		"  ____\n" +
		" |  _ \\\n" +
		" | |_) |   ___    __ _    ___    ___    _ __\n" +
		" |  _ <   / _ \\  / _` |  / __|  / _ \\  | '_ \\\n" +
		" | |_) | |  __/ | (_| | | (__  | (_) | | | | |\n" +
		" |____/   \\___|  \\__,_|  \\___|  \\___/  |_| |_|\n" +
		"=====================debug====================\n")
	clientTcp := config.String("client_tcp", "127.0.0.1:8000")
	TcpClient(clientTcp)
}

func Broadcast(bytes []byte) {
	echo.Print(bytes, runType)
	for _, c := range ClientMap {
		if err := c.WriteMsg(bytes, 0); err != nil {
			continue
		}
	}
}

func UdpServer(addr string) {
	uAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		panic(err)
	}
	client, err := net.ListenUDP("udp", uAddr)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	data := make([]byte, 1024*1024)
	for {
		n, err := client.Read(data)
		if err != nil {
			log.Println(err)
			continue
		}
		temp := data[0:n]
		log.Println(string(temp))
		Broadcast(temp)
	}
}

func TcpServer(addr string) {
	listener, err := conn.TcpListen(addr, "ctl")
	if err != nil {
		panic(err)
	}
	for c := range listener.Clients {
		go onConnect(c)
	}
}

func onConnect(client *conn.TcpConn) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("tunnelListener failed with error %v: %s", r, debug.Stack())
		}
	}()
	client.SetReadDeadline(time.Now().Add(10 * time.Second))
	//登录校验
	msg, typ, err := client.ReadMsg()
	if err != nil {
		client.Close()
		return
	}
	if typ != 3 {
		client.Close()
		return
	}
	if string(msg) == "" || string(msg) != config.String("password", "") {
		client.Close()
		return
	}
	client.SetReadDeadline(time.Time{})
	//客户端进入
	log.Printf("client income")
	ClientMap[client.Id] = client
	//循环监听
	for {
		_, typ, err = client.ReadMsg()
		if err != nil {
			delete(ClientMap, client.Id)
			client.Close()
			return
		}
		//回复心跳--
		if typ == 1 {
			if err = client.WriteMsg(nil, 2); err != nil {
				log.Printf("client quit:" + err.Error())
				continue
			}
		}
	}
}

func TcpClient(addr string) {
	client, err := conn.TcpDial(addr)
	if err != nil {
		log.Println("client exit:" + err.Error())
		return
	}

	log.Printf("New connection from %v", client.RemoteAddr())
	//发送登录密码
	password := config.String("password", "")
	if err = client.WriteMsg([]byte(password), 3); err != nil {
		log.Printf("client quit:" + err.Error())
		return
	}

	//发生心跳
	go func() {
		for {
			if err = client.WriteMsg(nil, 1); err != nil {
				break
			}
			time.Sleep(30 * time.Second)
		}
	}()
	for {
		rawMsg, typ, err := client.ReadMsg()
		if err != nil {
			log.Println(err)
			client.Close()
			return
		}
		//心跳包
		if typ == 0 {
			echo.Print(rawMsg, runType)
		}
	}
}
