package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/codedust/go-tox"
)

func sendFile(t *gotox.Tox, friendNumber uint32, fileNameOnDisc, fileNameToSend string) {
	file, err := os.Open(fileNameOnDisc)
	if err != nil {
		t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "File not found. Sorry!")
		file.Close()
		return
	}

	// get the file size
	stat, err := file.Stat()
	if err != nil {
		t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Could not read file stats.")
		file.Close()
		return
	}

	fmt.Println("File size is ", stat.Size())

	fileNumber, err := t.FileSend(friendNumber, gotox.TOX_FILE_KIND_DATA, uint64(stat.Size()), nil, fileNameToSend)
	if err != nil {
		t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "t.FileSend() failed.")
		file.Close()
		return
	}

	for _, t2 := range toxes {
		if t == t2.tox {
			if t2.users[friendNumber].transfers == nil {
				user2 := t2.users[friendNumber]
				user2.transfers = make(map[uint32]FileTransfer)
				t2.users[friendNumber] = user2
			}
			t2.users[friendNumber].transfers[fileNumber] = FileTransfer{fileHandle: file, fileSize: uint64(stat.Size())}
		}
	}
}

func onFileRecv(t *gotox.Tox, friendNumber uint32, fileNumber uint32, kind gotox.ToxFileKind, filesize uint64, filename string) {
	if kind == gotox.TOX_FILE_KIND_AVATAR {

		if filesize > MaxAvatarSize {
			// reject file send request
			t.FileControl(friendNumber, fileNumber, gotox.TOX_FILE_CONTROL_CANCEL)
			return
		}

		publicKey, _ := t.FriendGetPublickey(friendNumber)
		file, err := os.Create("files" + string(os.PathSeparator) + "avatar_" + hex.EncodeToString(publicKey) + ".png")
		if err != nil {
			fmt.Println("[ERROR] Error creating file", "avatar_"+hex.EncodeToString(publicKey)+".png")
		}

		// append the file to the map of active file transfers
		for _, t2 := range toxes {
			if t == t2.tox {
				if t2.users[friendNumber].transfers == nil {
					user2 := t2.users[friendNumber]
					user2.transfers = make(map[uint32]FileTransfer)
					t2.users[friendNumber] = user2
				}
				t2.users[friendNumber].transfers[fileNumber] = FileTransfer{fileHandle: file, fileSize: filesize}
			}
		}

		// accept the file send request
		t.FileControl(friendNumber, fileNumber, gotox.TOX_FILE_CONTROL_RESUME)

	} else {
		// accept files of any length

		name, err := t.FriendGetName(friendNumber)
		if err != nil {
			name = "unknown"
		}

		pub, err := t.FriendGetPublickey(friendNumber)
		var pubstr string
		if err == nil && len(pub) > 5 {
			pubstr = hex.EncodeToString(pub)
			pubstr = pubstr[:5]
		}

		file, err := os.Create("files" + string(os.PathSeparator) + name + "_" + pubstr + "_" + filename)
		if err != nil {
			fmt.Println("[ERROR] Error creating file", name+"_"+pubstr+"_"+filename)
		}

		// append the file to the map of active file transfers
		for _, t2 := range toxes {
			if t == t2.tox {
				if t2.users[friendNumber].transfers == nil {
					user2 := t2.users[friendNumber]
					user2.transfers = make(map[uint32]FileTransfer)
					t2.users[friendNumber] = user2
				}
				t2.users[friendNumber].transfers[fileNumber] = FileTransfer{fileHandle: file, fileSize: filesize}
			}
		}

		// accept the file send request
		t.FileControl(friendNumber, fileNumber, gotox.TOX_FILE_CONTROL_RESUME)
	}
}

func onFileRecvControl(t *gotox.Tox, friendNumber uint32, fileNumber uint32, fileControl gotox.ToxFileControl) {
	var transfer FileTransfer
	var ok bool
	for _, t2 := range toxes {
		if t == t2.tox {
			if t2.users[friendNumber].transfers == nil {
				user2 := t2.users[friendNumber]
				user2.transfers = make(map[uint32]FileTransfer)
				t2.users[friendNumber] = user2
			}
			transfer, ok = t2.users[friendNumber].transfers[fileNumber]
		}
	}

	if !ok {
		fmt.Println("Error: File handle does not exist")
		return
	}

	if fileControl == gotox.TOX_FILE_CONTROL_CANCEL {
		// delete file handle
		numerofopenedhandlers := 0
		for _, t2 := range toxes {
			if t == t2.tox {
				for k, user := range t2.users {
					if t2.users[k].transfers == nil {
						user2 := t2.users[k]
						user2.transfers = make(map[uint32]FileTransfer)
						t2.users[k] = user2
					}
					for _, transfer2 := range user.transfers {
						if transfer.fileHandle == transfer2.fileHandle {
							numerofopenedhandlers++
						}
					}
				}
			}
		}
		//fmt.Println("test1: " + strconv.Itoa(numerofopenedhandlers))
		if numerofopenedhandlers <= 1 {
			transfer.fileHandle.Sync()
			transfer.fileHandle.Close()
		}
		for _, t2 := range toxes {
			if t == t2.tox {
				delete(t2.users[friendNumber].transfers, fileNumber)
			}
		}

	}
}

func onFileChunkRequest(t *gotox.Tox, friendNumber uint32, fileNumber uint32, position uint64, length uint64) {
	var transfer FileTransfer
	var ok bool
	for _, t2 := range toxes {
		if t == t2.tox {
			if t2.users[friendNumber].transfers == nil {
				user2 := t2.users[friendNumber]
				user2.transfers = make(map[uint32]FileTransfer)
				t2.users[friendNumber] = user2
			}
			transfer, ok = t2.users[friendNumber].transfers[fileNumber]
		}
	}
	if !ok {
		fmt.Println("Error: File handle does not exist")
		return
	}

	// a zero-length chunk request confirms that the file was successfully transferred
	if length == 0 {
		numerofopenedhandlers := 0
		for _, t2 := range toxes {
			if t == t2.tox {
				for k, user := range t2.users {
					if t2.users[k].transfers == nil {
						user2 := t2.users[k]
						user2.transfers = make(map[uint32]FileTransfer)
						t2.users[k] = user2
					}
					for _, transfer2 := range user.transfers {
						if transfer.fileHandle == transfer2.fileHandle {
							numerofopenedhandlers++
						}
					}
				}
			}
		}
		//fmt.Println("test2: " + strconv.Itoa(numerofopenedhandlers))
		if numerofopenedhandlers <= 1 {
			transfer.fileHandle.Close()
		}
		for _, t2 := range toxes {
			if t == t2.tox {
				delete(t2.users[friendNumber].transfers, fileNumber)
			}
		}
		fmt.Println("File transfer completed (sending)", fileNumber)
		return
	}

	// read the requested data to send
	data := make([]byte, length)

	for _, t2 := range toxes {
		if t == t2.tox {
			_, err := t2.users[friendNumber].transfers[fileNumber].fileHandle.ReadAt(data, int64(position))
			if err != nil {
				fmt.Println("Error reading file", err)
				return
			}
		}
	}

	// send the requested data
	t.FileSendChunk(friendNumber, fileNumber, position, data)
}

func onFileRecvChunk(t *gotox.Tox, friendNumber uint32, fileNumber uint32, position uint64, data []byte) {
	var transfer FileTransfer
	var ok bool
	for _, t2 := range toxes {
		if t == t2.tox {
			if t2.users[friendNumber].transfers == nil {
				user2 := t2.users[friendNumber]
				user2.transfers = make(map[uint32]FileTransfer)
				t2.users[friendNumber] = user2
			}
			transfer, ok = t2.users[friendNumber].transfers[fileNumber]
		}
	}
	if !ok {
		if len(data) == 0 {
			// ignore the zero-length chunk that indicates that the transfer is
			// complete (see below)
			return
		}

		fmt.Println("Error: File handle does not exist")
		return
	}

	// write the received data to the file handle
	transfer.fileHandle.WriteAt(data, (int64)(position))

	// file transfer completed
	if position+uint64(len(data)) >= transfer.fileSize {
		// Some clients will send us another zero-length chunk without data (only
		// required for streams, not necessary for files with a known size) and some
		// will not.
		// We will delete the file handle now (we aleady received the whole file)
		// and ignore the file handle error when the zero-length chunk arrives.

		filename := transfer.fileHandle.Name()

		numerofopenedhandlers := 0
		for _, t2 := range toxes {
			if t == t2.tox {
				for k, user := range t2.users {
					if t2.users[k].transfers == nil {
						user2 := t2.users[k]
						user2.transfers = make(map[uint32]FileTransfer)
						t2.users[k] = user2
					}
					for _, transfer2 := range user.transfers {
						if transfer.fileHandle == transfer2.fileHandle {
							numerofopenedhandlers++
						}
					}
				}
			}
		}
		//fmt.Println("test3: " + strconv.Itoa(numerofopenedhandlers))
		if numerofopenedhandlers <= 1 {
			transfer.fileHandle.Sync()
			transfer.fileHandle.Close()
		}
		for _, t2 := range toxes {
			if t == t2.tox {
				delete(t2.users[friendNumber].transfers, fileNumber)
			}
		}
		fmt.Println("File transfer completed (receiving)", fileNumber)

		if !(len(filename) > 13 && filename[:13] == "files/avatar_") {
			file, err := os.Open(filename)
			if err != nil {
				t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Error sending a file: File not found.'")
				file.Close()
				return
			}

			// get the file size
			stat, err := file.Stat()
			if err != nil {
				t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Error sending a file: Could not read file stats.")
				file.Close()
				return
			}

			fmt.Println("File size is ", stat.Size())

			//normalize name
			if len(filename) > 6 && filename[:6] == "files/" {
				filename = filename[6:]
			}

			allFriends, err := t.SelfGetFriendlist()
			if err != nil {
				return
			}

			for _, v := range allFriends {
				if v == friendNumber {
					continue
				}

				friendStatus, err := t.FriendGetConnectionStatus(v)
				if err != nil {
					continue
				}

				if friendStatus != gotox.TOX_CONNECTION_NONE {
					fileNumber, err := t.FileSend(v, gotox.TOX_FILE_KIND_DATA, uint64(stat.Size()), nil, filename)
					if err != nil {
						t.FriendSendMessage(v, gotox.TOX_MESSAGE_TYPE_ACTION, "Error sending a file "+filename+": t.FileSend() failed.")
						file.Close()
						return
					}

					for _, t2 := range toxes {
						if t == t2.tox {
							if t2.users[v].transfers == nil {
								user2 := t2.users[v]
								user2.transfers = make(map[uint32]FileTransfer)
								t2.users[v] = user2
							}
							t2.users[v].transfers[fileNumber] = FileTransfer{fileHandle: file, fileSize: uint64(stat.Size())}
						}
					}
				}
			}

			for k := range toxes {
				if t == toxes[k].tox {
					toxes[k].lastMessageFrom = 4294967295 //make the >>> From: appear for the next message
				}
			}

			t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "File send!")
		} else {
			//t.FriendSendMessage(friendNumber, gotox.TOX_MESSAGE_TYPE_ACTION, "Welcome!")
		}
	}
}

// loadData reads a file and returns its content as a byte array
func loadData(filepath string) ([]byte, error) {
	if len(filepath) == 0 {
		return nil, errors.New("Empty path")
	}

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return data, err
}

// saveData writes the savedata from toxcore to a file
func saveData(t *gotox.Tox, filepath string) error {
	if len(filepath) == 0 {
		return errors.New("Empty path")
	}

	data, err := t.GetSavedata()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath, data, 0644)
	return err
}
