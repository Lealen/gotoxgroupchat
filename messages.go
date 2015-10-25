package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/codedust/go-tox"
)

func sendToAllFriends(t *gotox.Tox, sendas uint32, message string) error {
	lastmessageTime := time.Now()
	for k, t2 := range toxes {
		if t == t2.tox {
			lastmessageTime = t2.users[sendas].lastMessage.messageTime
			if lastmessageTime.After(time.Now().Add(-10*time.Second)) && message == t2.users[sendas].lastMessage.text {
				fmt.Println("Message canceled, too quick")
				return nil
			}
			usr := t2.users[sendas]
			usr.lastMessage.messageTime = time.Now()
			usr.lastMessage.text = message
			toxes[k].users[sendas] = usr

			if t2.lastMessageFrom != sendas || lastmessageTime.Before(time.Now().Add(-3*time.Minute)) {
				//t.FriendSendMessage(v, gotox.TOX_MESSAGE_TYPE_ACTION, "From: "+name)
				message = ">>> From: " + getFriendName(t, sendas) + "\n" + message
			}
			toxes[k].lastMessageFrom = sendas
		}
	}

	allFriends, err := t.SelfGetFriendlist()
	if err != nil {
		return err
	}

	searchforurls := strings.Split(message, " ")
	searchforurlsnumber := 0
	for _, v := range searchforurls {
		//fmt.Println(v)
		if searchforurlsnumber > 2 {
			break
		}
		if len(v) > 10 && (v[:7] == "http://" || v[:8] == "https://") {
			link, err := url.Parse(v)
			if err != nil {
				log.Print(err)
				continue
			}
			searchforurlsnumber++
			resp, err := httpclient.Get(link.String())
			if err != nil || resp.StatusCode != 200 {
				log.Print(err)
				continue
			}
			html, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Print(err)
				continue
			}
			title := httptitleregexp.FindString(string(html))
			if len(title) > 1 {
				if len(title) > 7 && title[:7] == "<title>" {
					title = title[7:]
				}
				if len(title) > 8 && title[len(title)-8:] == "</title>" {
					title = title[:len(title)-8]
				}
				//searchforurls[k] = searchforurls[k] + "\n↪ " + title + "\n"
				title = strings.TrimSpace(title)
				message = message + "\n↪ " + title
				t.FriendSendMessage(sendas, gotox.TOX_MESSAGE_TYPE_ACTION, "↪ "+title)
			}
			//log.Print("test: " + title)
		}
	}
	//message = strings.Join(searchforurls, " ")

	for _, v := range allFriends {
		friendStatus, err := t.FriendGetConnectionStatus(v)
		if err != nil {
			continue
		}

		if v == sendas {
			continue
		}

		if friendStatus != gotox.TOX_CONNECTION_NONE {
			t.FriendSendMessage(v, gotox.TOX_MESSAGE_TYPE_NORMAL, message)
		}
	}

	return nil
}

func sendServerMessageToEveryone(t *gotox.Tox, message string) error {
	allFriends, err := t.SelfGetFriendlist()
	if err != nil {
		return err
	}

	for _, v := range allFriends {
		friendStatus, err := t.FriendGetConnectionStatus(v)
		if err != nil {
			continue
		}

		if friendStatus != gotox.TOX_CONNECTION_NONE {
			t.FriendSendMessage(v, gotox.TOX_MESSAGE_TYPE_ACTION, message)
		}
	}
	return nil
}

func onFriendRequest(t *gotox.Tox, publicKey []byte, message string) {
	fmt.Printf("New friend request from %s\n", hex.EncodeToString(publicKey))
	fmt.Printf("With message: %v\n", message)
	// Auto-accept friend request
	t.FriendAddNorequest(publicKey)
}

func onFriendMessage(t *gotox.Tox, friendNumber uint32, messagetype gotox.ToxMessageType, message string) {
	if messagetype == gotox.TOX_MESSAGE_TYPE_NORMAL {
		fmt.Printf("New message from %d : %s\n", friendNumber, message)
	} else {
		fmt.Printf("New action from %d : %s\n", friendNumber, message)
	}

	if len(message) > 1 && message[:1] == "/" || len(message) > 2 && message[:2] == "--" {
		if len(message) > 1 && message[:1] == "/" {
			message = message[1:]
		} else if len(message) > 2 && message[:2] == "--" {
			message = message[2:]
		}
		var command = message
		var args []string
		if strings.Index(command, " ") != -1 {
			tmp := strings.SplitN(command, " ", 2)
			command = strings.ToLower(tmp[0])
			args = strings.Split(tmp[1], " ")
		}
		switch command {
		case "online":
			allFriends, err := t.SelfGetFriendlist()
			if err != nil {
				return
			}
			var allOnline []string
			for _, v := range allFriends {
				friendStatus, err := t.FriendGetConnectionStatus(v)
				if err != nil {
					continue
				}

				if friendStatus != gotox.TOX_CONNECTION_NONE {
					name, err := t.FriendGetName(v)
					if err != nil {
						name = "unknown"
					}

					pub, err := t.FriendGetPublickey(v)
					var pubstr string
					if err == nil && len(pub) > 5 {
						pubstr = hex.EncodeToString(pub)
						pubstr = pubstr[:5]
					}

					allOnline = append(allOnline, name+" ("+pubstr+"...)")
				}
			}
			sort.Strings(allOnline)

			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Online: "+strings.Join(allOnline, ", "))
		case "users":
			allFriends, err := t.SelfGetFriendlist()
			if err != nil {
				return
			}
			var allUsers []string
			for _, v := range allFriends {
				name, err := t.FriendGetName(v)
				if err != nil {
					name = "unknown"
				}

				pub, err := t.FriendGetPublickey(v)
				var pubstr string
				if err == nil && len(pub) > 5 {
					pubstr = hex.EncodeToString(pub)
					pubstr = pubstr[:5]
				}

				allUsers = append(allUsers, name+" ("+pubstr+"...)")
			}
			sort.Strings(allUsers)

			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Users: "+strings.Join(allUsers, ", "))
		case "credits":
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Script by Lealen bez\nThis script will not occur without the participation of:\ndd (da313...)\nM (6eca7...)\ntm (60740...)")
		case "version":
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "gotoxgroupchat v0.7.10.25-r2 by Lealen bez\nGNU Terry Pratchett")
		case "unstuck":
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Teleporting...\nWHOOSH!")
		case "moo":
			sendFile(t, friendNumber, "moo.png", "moo.png")
		case "anonymoose":
			sendFile(t, friendNumber, "anonymoose.png", "anonymoose.png")
		case "why":
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "why not")
		case "why?":
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "why not?")
		case "answer":
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "42")
		case "q", "quote", "motd":
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, motdslice[random.Int()%len(motdslice)])
		case "test":
			t.SelfSetName(" ")
		case "help":
			//t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_NORMAL, "Type '/file' to receive a file.")
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "/anonymoose /id /moo /motd /online /ping /pong /roll /q /quote /unstuck /users /version /why /why?")
		case "ping":
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "pong")
		case "pong":
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "ping")
		case "roll":
			roll(t, friendNumber, args)
		case "id":
			cmdGetID(t, friendNumber, args)
		case "adminsendmessage":
			cmdAdminSendMessage(t, friendNumber, args)
		case "adminsendfile":
			cmdAdminSendFile(t, friendNumber, args)
		default:
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Command not found.\nType '/help' for available commands.")
		}
	} else {
		switch message {
		case "ping":
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "pong")
		case "pong":
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "ping")
		}

		sendToAllFriends(t, friendNumber, message)
	}

}
