package main

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/codedust/go-tox"
)

func roll(t *gotox.Tox, friendNumber uint32, args []string) {
	if len(args) == 0 {
		t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "You need to supply at least one argument!\nExample: /roll 1d6+10")
		return
	} else if len(args) > 3 {
		t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Too many arguments")
		return
	}
	var text = ""
	for _, v := range args {
		if text != "" {
			text += "\n"
		}
		var diceresults []string
		result := 0
		dice := v
		operation := ""
		additional := "0"
		if strings.Index(v, "+") != -1 {
			tmp := strings.SplitN(v, "+", 2)
			dice = tmp[0]
			operation = "+"
			additional = tmp[1]
		} else if strings.Index(v, "-") != -1 {
			tmp := strings.SplitN(v, "-", 2)
			dice = tmp[0]
			operation = "-"
			additional = tmp[1]
		}

		if strings.Index(dice, "d") != -1 {
			tmp := strings.SplitN(dice, "d", 2)
			if len(tmp) != 2 {
				t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Wrong dice type: "+v)
				return
			}
			i, err := strconv.Atoi(tmp[0])
			if err != nil || i <= 0 {
				t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Wrong dice type: "+v)
				return
			}
			if i > 123 {
				t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Wrong dice type: "+v+"; Why "+tmp[0]+"?")
				return
			}
			d, err := strconv.Atoi(tmp[1])
			if err != nil || d <= 0 {
				t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Wrong dice type: "+v)
				return
			}
			if d > 100000 {
				t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Wrong dice type: "+v+"; Why "+tmp[1]+"?")
				return
			}
			for ; i > 0; i-- {
				rnd := random.Int()%d + 1
				diceresults = append(diceresults, strconv.Itoa(rnd))
				result += rnd
			}
		} else {
			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Wrong dice type: "+v)
			return
		}

		if len(diceresults) == 1 && additional == "0" {
			text += getFriendName(t, friendNumber) + " rolls the " + v + " dice, result: " + strings.Join(diceresults, ", ")
		} else {
			text += getFriendName(t, friendNumber) + " rolls the " + v + " dice, results: " + strings.Join(diceresults, ", ")
			if additional != "0" {
				text += ", " + operation + additional
				addit, err := strconv.Atoi(additional)
				if err != nil {
					t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Wrong dice type: "+v)
					return
				}
				if addit > 1000000000 {
					t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Wrong dice type: "+v+"; Why "+additional+"?")
					return
				}
				if operation == "+" {
					result += addit
				} else if operation == "-" {
					result -= addit
				} else {
					t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Wrong dice type: "+v)
					return
				}
			}

			text += "; Total: " + strconv.Itoa(result)
		}
	}
	sendServerMessageToEveryone(t, text)
}

func cmdGetID(t *gotox.Tox, friendNumber uint32, args []string) {
	if len(args) == 0 {
		t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "You need to supply at least one argument!\nExample: /id user or /id PuBkEy")
		return
	} else if len(args) > 1 {
		t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Too many arguments")
		return
	}

	userNumber := searchForUser(t, args[0])

	pubkey, err := t.FriendGetPublickey(userNumber)
	if err != nil {
		t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Error getting ID tox.FriendGetPublickey() failed")
		return
	}

	t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, getFriendName(t, userNumber)+" ID: "+hex.EncodeToString(pubkey))
}

func cmdAdminSendMessage(t *gotox.Tox, friendNumber uint32, args []string) {
	pub, err := t.FriendGetPublickey(friendNumber)
	if err != nil {
		return
	}
	pubstr := hex.EncodeToString(pub)

	if pubstr != "3217fd65a96e38c9334f209d96643f0f9b091b37d0e7dd9df128592ee29e6c6f" {
		t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Tu straż! Otwórzcie drzwi! W przeciwnym razie Detrytus je otworzy. A kiedy on otwiera drzwi, to już zostają otwarte.")
		return
	}

	if len(args) == 0 {
		t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "You need to supply at least one argument!\nExample: /id user or /id PuBkEy")
		return
	}

	sendServerMessageToEveryone(t, strings.Join(args, " "))
}

func cmdAdminSendFile(t *gotox.Tox, friendNumber uint32, args []string) {
	pub, err := t.FriendGetPublickey(friendNumber)
	if err != nil {
		return
	}
	pubstr := hex.EncodeToString(pub)

	if pubstr != "3217fd65a96e38c9334f209d96643f0f9b091b37d0e7dd9df128592ee29e6c6f" {
		t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Tu straż! Otwórzcie drzwi! W przeciwnym razie Detrytus je otworzy. A kiedy on otwiera drzwi, to już zostają otwarte.")
		return
	}

	if len(args) == 0 {
		t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "You need to supply at least one argument!\nExample: /id user or /id PuBkEy")
		return
	}

	fileNameToSend := args[0]

	//normalize name
	if len(fileNameToSend) > 6 && fileNameToSend[:6] == "files/" {
		fileNameToSend = fileNameToSend[6:]
	}

	if len(args) > 1 {
		fileNameToSend = strings.Join(args[1:], " ")
	}

	allFriends, err := t.SelfGetFriendlist()
	if err != nil {
		return
	}
	for _, v := range allFriends {
		friendStatus, err := t.FriendGetConnectionStatus(v)
		if err != nil {
			continue
		}

		if friendStatus != gotox.TOX_CONNECTION_NONE {
			sendFile(t, v, args[0], fileNameToSend)
		}
	}
}
