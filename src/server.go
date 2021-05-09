package main

import (
	"fmt"
	"log"
	"net"

	proto "google.golang.org/protobuf/proto"
)

var timeline []*Post

func server() {
	// Main server that runs
	addr, e := net.ResolveUDPAddr("udp", ":"+port)
	handleError(e, "Server Resolve")

	connection, e := net.ListenUDP("udp", addr)
	handleError(e, "Server Listen")
	defer connection.Close()

	buf := make([]byte, 1024*125)

	for {
		//listen for ACK
		numBytes, _, e := connection.ReadFromUDP(buf)
		handleError(e, "Server Read")
		messagePacket := &Message{}
		e = proto.Unmarshal(buf[0:numBytes], messagePacket)
		handleError(e, "Server Proto Unmarshal")

		log.Println("Received packet of type: ", messagePacket.GetType())
		switch msgType := messagePacket.GetType(); msgType {
		case 0: // view posts
			refreshView(messagePacket.GetPosts())
		case 1: // post request
			resendPost(messagePacket.GetOptional())
		}
	}
}

func refreshView(posts []*Post) {
	timeline = posts
	for _, post := range posts {
		log.Println(post.GetTtl(), getTime())
		fmt.Println(getUserFromPost(post), ": ", post.GetData())
	}
	log.Println("Finished view request at :", getMilliseconds())
}

func resendPost(postID string) {
	post := myPosts[postID]
	post.Ttl = getTime() + 86400

	posts := make([]*Post, 0)
	posts = append(posts, post)

	messagePacket := createMessage(5, "", posts)
	sendMessage(messagePacket, csUDP)
}
