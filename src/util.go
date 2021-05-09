package main

import (
	"log"
	"net"
	"strings"
	"time"
)

func getTime() uint64 {
	return uint64(time.Now().Unix())
}

func getMilliseconds() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func getUserFromPost(post *Post) string {
	return strings.Split(post.GetId(), "&")[0]
}

func handleError(e error, custom string) {
	if e != nil {
		log.Println(custom)
		log.Fatal(e)
	}
}

func getLocalIP() string {
	//get own IP addr, can dial any random IP to get outbound
	conn, e := net.Dial("udp", "8.8.8.8:80")
	handleError(e, "getLocalIP")

	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
