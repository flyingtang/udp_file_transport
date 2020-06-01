package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {

	var (
		ip   string
		port string

		senDir  string
		recvDic string
	)

	flag.StringVar(&ip, "ip", "", "remote listen ip")
	flag.StringVar(&port, "port", "", "remote listen port and local listen port")
	flag.StringVar(&senDir, "sd", "/geesunn/transport/j212_send", "send directory")
	flag.StringVar(&recvDic, "rd", "/geesunn/transport/j212_recv", "receive directory")

	flag.Parse()
	//TODO check parameters

	// 监听接收数据
	go func() {

		defer func() {
			if r := recover(); r != nil {
				log.Print(r)
			}
		}()

		ipAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%v", port))
		if err != nil {
			log.Fatal(err)
			return
		}

		conn, err := net.ListenUDP("udp", ipAddr)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer conn.Close()

		buf := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Print(err)
			}

			log.Print("读到得数据", buf[:n])
		}
	}()

	// 发送本地数据
	go func() {
		ipAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", ip, port))
		if err != nil {
			log.Fatal(err)
			return
		}

		conn, err := net.DialUDP("udp", nil, ipAddr)
		if err != nil {
			log.Fatal(err)
			return
		}

		for {
			if fs, err := ioutil.ReadDir(senDir); err != nil {
				log.Print(err)
				continue
			} else {
				for _, f := range fs {
					p := filepath.Join(senDir, f.Name())
					fi, err := os.Open(p)
					if err != nil {
						log.Print(err)
						continue
					}
					if _, err := io.Copy(conn, fi); err != nil {
						log.Print(err)
						continue
					}
				}
			}
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	os.Exit(0)
}

type SyncContent struct {
	UUID string `json:"uuid"`
	Type string `json:"type" description: "数据类型 http_req.db http_resp.db redis.db"`
	From string `json:"from"`
	To   string `json:"to"`
	Data string `json:"data"`
	//Unixtime int64  `json:"unixtime"`
}
