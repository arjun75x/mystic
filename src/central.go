package main

import (
	"log"
	"net"
	sync "sync"

	proto "google.golang.org/protobuf/proto"
)

var userToIP = make(map[string]string)
var userToConn = make(map[string]*net.UDPConn)
var userToFollowing = make(map[string]*set)
var userToPost = make(map[string]map[string]*Post)

var wgp *sync.WaitGroup

func centralServer() {
	// Main server that runs
	addr, e := net.ResolveUDPAddr("udp", ":"+port)
	handleError(e, "Server Resolve")

	connection, e := net.ListenUDP("udp", addr)
	handleError(e, "Server Listen")
	defer connection.Close()

	buf := make([]byte, 1024)

	for {
		//listen for ACK
		numBytes, udpaddr, e := connection.ReadFromUDP(buf)
		handleError(e, "Server Read")
		messagePacket := &Message{}
		e = proto.Unmarshal(buf[0:numBytes], messagePacket)
		handleError(e, "Server Proto Unmarshal")

		log.Println("Received packet of type: ", messagePacket.GetType())
		switch msgType := messagePacket.GetType(); msgType {
		case 0: // join
			handleJoin(udpaddr.IP.String(), messagePacket.GetUsername())
		case 1: // follow
			handleFollow(messagePacket.GetUsername(), messagePacket.GetOptional())
		case 2: // unfollow
			handleUnfollow(messagePacket.GetUsername(), messagePacket.GetOptional())
		case 3: // post
			handlePost(messagePacket.GetUsername(), messagePacket.GetPosts()[0])
		case 4: // view
			go handleView(messagePacket.GetUsername())
		case 5: // resent post
			handleResentPost(messagePacket.GetUsername(), messagePacket.GetPosts()[0])
		}
	}
}

func handleJoin(ipAddr string, username string) {
	userToIP[username] = ipAddr

	serverAddr, err := net.ResolveUDPAddr("udp", ipAddr+":"+port)
	handleError(err, "Client Resolve")

	udpconn, err := net.DialUDP("udp", nil, serverAddr)
	handleError(err, "Client Dial")
	userToConn[username] = udpconn

	userToFollowing[username] = NewSet()

	userToPost[username] = make(map[string]*Post)
}

func handleFollow(user string, other string) {
	userToFollowing[user].Add(other)
}

func handleUnfollow(user string, other string) {
	userToFollowing[user].Remove(other)
}

func handlePost(user string, post *Post) {
	userToPost[user][post.GetId()] = post
}

func handleResentPost(user string, post *Post) {
	if wgp != nil {
		defer wgp.Done()
	}
	userToPost[user][post.GetId()] = post
}

// TODO: Fix timeline logic to actually get most recent posts instead of top 5 per user
func handleView(user string) {
	wg := sync.WaitGroup{}
	wgp = &wg
	out := make([]*Post, 0)

	for _, other := range userToFollowing[user].List() {
		i := 0
		for _, post := range userToPost[other] {
			// if i > 5 {
			// 	break
			// }
			if post.GetTtl() == 0 {
				wg.Add(1)
				rerequestPost(post)
			}
			out = append(out, post)
			// if len(out) >= 50 {
			// 	goto End
			// }
			i += 1
		}
	}

	// End:
	wg.Wait()
	for i, post := range out {
		user := getUserFromPost(post)
		newPost := userToPost[user][post.GetId()]
		if post.GetTtl() == 0 && newPost.GetTtl() != 0 {
			out[i] = newPost
		}
	}

	sendPostsToClient(user, out)
}

func rerequestPost(post *Post) {
	log.Println("Rerequest post ", post.GetId())
	user := getUserFromPost(post)
	conn := userToConn[user]
	messagePacket := createMessage(1, post.GetId(), make([]*Post, 0))
	sendMessage(messagePacket, conn)
}

func sendPostsToClient(user string, posts []*Post) {
	message := Message{
		Type:  0,
		Posts: posts,
	}

	serverAddr, err := net.ResolveUDPAddr("udp", userToIP[user]+":"+port)
	handleError(err, "Client Resolve")

	udpconn, err := net.DialUDP("udp", nil, serverAddr)
	handleError(err, "Client Dial")

	out, err := proto.Marshal(&message)
	if err != nil {
		log.Fatalln("Failed to encode message packet:", err)
	}
	sendBuf := []byte(out)
	log.Println("Sent view")
	_, err = udpconn.Write(sendBuf)
	if err != nil {
		log.Println("Write error", err)
	}
}

func ttlCheck() {
	for {
		time := getTime()
		for _, posts := range userToPost {
			for _, post := range posts {
				if time > post.GetTtl() && post.GetTtl() > 0 {
					log.Println("TTL expired on ", post.GetId())
					post.Ttl = 0
					post.Data = ""
				}
			}
		}
	}
}
