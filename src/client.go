package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	proto "google.golang.org/protobuf/proto"
)

var csUDP *net.UDPConn

var myPosts = make(map[string]*Post) // keep track of my posts
var globalID = 0

func join() {
	openCSconnection()
	messagePacket := createMessage(0, "", make([]*Post, 0))
	sendMessage(messagePacket, csUDP)
}

func follow(other string) {
	messagePacket := createMessage(1, other, make([]*Post, 0))
	sendMessage(messagePacket, csUDP)
}

func unfollow(other string) {
	messagePacket := createMessage(2, other, make([]*Post, 0))
	sendMessage(messagePacket, csUDP)
}

func post(ttl string, data string) {
	val, _ := strconv.Atoi(ttl)
	post := Post{
		Id:   username + "&" + strconv.Itoa(globalID),
		Ttl:  getTime() + uint64(val),
		Data: data,
	}
	globalID += 1
	myPosts[post.GetId()] = &post

	posts := make([]*Post, 0)
	posts = append(posts, &post)

	messagePacket := createMessage(3, "", posts)
	sendMessage(messagePacket, csUDP)
}

func view() {
	log.Println("Sent view request at :", getMilliseconds())
	messagePacket := createMessage(4, "", make([]*Post, 0))
	sendMessage(messagePacket, csUDP)
}

func test(num_posts string, ttl string) {
	n, _ := strconv.Atoi(num_posts)

	for i := 0; i < n; i++ {
		post(ttl, fmt.Sprintf(username, "_rand_post_", i))
	}
}

func openCSconnection() {
	serverAddr, err := net.ResolveUDPAddr("udp", centralIP+":"+port)
	handleError(err, "Client Resolve")

	udpconn, err := net.DialUDP("udp", nil, serverAddr)
	handleError(err, "Client Dial")

	csUDP = udpconn
}

func sendMessage(messagePacket *Message, conn *net.UDPConn) {
	out, err := proto.Marshal(messagePacket)
	if err != nil {
		log.Fatalln("Failed to encode message packet:", err)
	}
	sendBuf := []byte(out)
	log.Println("Sent message")
	_, err = conn.Write(sendBuf)
	if err != nil {
		log.Println("Write error", err)
	}
}

func createMessage(type_ uint32, other string, posts []*Post) *Message {
	message := Message{
		Type:     type_,
		Username: username,
		Posts:    posts,
		Optional: other,
	}
	return &message
}
