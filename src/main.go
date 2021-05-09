package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var centralIP string = "172.22.158.235"
var port = "4000"
var username string
var (
	toPrint = flag.Bool("p", false, "Print or not")
)

func main() {
	flag.Parse()
	localIP := getLocalIP()
	scanner := bufio.NewScanner(os.Stdin)

	f, err := os.OpenFile(fmt.Sprint("logs/", getTime(), ".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	if *toPrint {
		wrt := io.MultiWriter(os.Stdout, f)
		log.SetOutput(wrt)
	} else {
		log.SetOutput(f)
	}

	if centralIP == localIP {
		log.Println("This is central server")
		go ttlCheck()
		centralServer()
	} else {
		// CLI
		go server()
		for {
			if scanner.Scan() {
				split := strings.Split(scanner.Text(), " ")
				switch split[0] {
				// commands
				case "join":
					username = split[1]
					join()
				case "follow":
					follow(split[1])
				case "unfollow":
					unfollow(split[1])
				case "post":
					post(split[1], strings.Join(split[2:], " "))
				case "view":
					view()
				case "test": // used for experiments
					test(split[1], split[2])
				}
			}
		}
	}
}
