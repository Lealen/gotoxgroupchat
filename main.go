package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"time"

	"github.com/codedust/go-tox"
)

type Server struct {
	Address   string
	Port      uint16
	PublicKey []byte
}

const MAX_AVATAR_SIZE = 65536 // see github.com/Tox/Tox-STS/blob/master/STS.md#avatars

type FileTransfer struct {
	fileHandle *os.File
	fileSize   uint64
}

var httpclient *http.Client
var httptitleregexp *regexp.Regexp

var randomSource = rand.NewSource(time.Now().UnixNano())
var random = rand.New(randomSource)

func init() {
	var err error
	httpclient = &http.Client{
		Timeout: time.Second * 3,
	}
	httptitleregexp, err = regexp.Compile("<title.*>(.*?)</title>")
	if err != nil {
		panic(err)
	}
}

type message struct {
	text        string
	messageTime time.Time
}

type user struct {
	// Map of active file transfers
	transfers   map[uint32]FileTransfer
	lastMessage message
}

type toxe struct {
	tox *gotox.Tox
	filepath,
	botname string
	displayUserInStatusText,
	addSonendofStatus bool
	users           map[uint32]user
	lastMessageFrom uint32
}

var toxes []toxe

func run(filepath, botname string, displayUserInStatusText, addSonendofStatus bool) {
	var newToxInstance bool = false
	var options *gotox.Options

	savedata, err := loadData(filepath)
	if err == nil {
		fmt.Println("[INFO] Loading Tox profile from savedata...")
		options = &gotox.Options{
			IPv6Enabled:  true,
			UDPEnabled:   true,
			ProxyType:    gotox.TOX_PROXY_TYPE_NONE,
			ProxyHost:    "127.0.0.1",
			ProxyPort:    5555,
			StartPort:    0,
			EndPort:      0,
			TcpPort:      0, // only enable TCP server if your client provides an option to disable it
			SaveDataType: gotox.TOX_SAVEDATA_TYPE_TOX_SAVE,
			SaveData:     savedata}
	} else {
		fmt.Println("[INFO] Creating new Tox profile...")
		options = nil // default options
		newToxInstance = true
	}

	tox, err := gotox.New(options)
	if err != nil {
		panic(err)
	}

	if newToxInstance {
		tox.SelfSetName(" ")
		tox.SelfSetStatusMessage("gotox is cool!")
	}

	addr, _ := tox.SelfGetAddress()
	fmt.Println("ID: ", hex.EncodeToString(addr))

	err = tox.SelfSetStatus(gotox.TOX_USERSTATUS_NONE)

	// Register our callbacks
	tox.CallbackFriendRequest(onFriendRequest)
	tox.CallbackFriendMessage(onFriendMessage)
	tox.CallbackFileRecvControl(onFileRecvControl)
	tox.CallbackFileChunkRequest(onFileChunkRequest)
	tox.CallbackFileRecv(onFileRecv)
	tox.CallbackFileRecvChunk(onFileRecvChunk)

	/* Connect to the network
	 * Use more than one node in a real world szenario. This example relies one
	 * the following node to be up.
	 */
	pubkey, _ := hex.DecodeString("04119E835DF3E78BACF0F84235B300546AF8B936F035185E2A8E9E0A67C8924F")
	server := &Server{"144.76.60.215", 33445, pubkey}

	err = tox.Bootstrap(server.Address, server.Port, server.PublicKey)
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			updateTyping(tox)
			updateStatus(tox)
		}
	}()

	toxes = append(toxes, toxe{tox, filepath, botname, displayUserInStatusText, addSonendofStatus, make(map[uint32]user), 4294967295})
}

func main() {

	//flag.StringVar(&filepath, "save", "./toxsavedata", "path to save file")
	//flag.Parse()

	fmt.Printf("[INFO] Using Tox version %d.%d.%d\n", gotox.VersionMajor(), gotox.VersionMinor(), gotox.VersionPatch())

	if !gotox.VersionIsCompatible(0, 0, 0) {
		fmt.Println("[ERROR] The compiled library (toxcore) is not compatible with this example.")
		fmt.Println("[ERROR] Please update your Tox library. If this error persists, please report it to the gotox developers.")
		fmt.Println("[ERROR] Thanks!")
		return
	}

	run("./toxsavedata", "gopher hacker", false, true)
	run("./toxsavedata2", "czwórka wspaniałych", false, false)

	isRunning := true

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ticker := time.NewTicker(25 * time.Millisecond)

	for isRunning {
		select {
		case <-c:
			fmt.Printf("\nSaving...\n")
			for _, t := range toxes {
				if err := saveData(t.tox, t.filepath); err != nil {
					fmt.Println("[ERROR]", err)
				}
			}
			fmt.Println("Killing")
			isRunning = false
			for _, t := range toxes {
				t.tox.Kill()
			}
		case <-ticker.C:
			for _, t := range toxes {
				t.tox.Iterate()
			}
		}
	}
}

func updateStatus(t *gotox.Tox) {
	allFriends, err := t.SelfGetFriendlist()
	if err != nil {
		return
	}

	i := 0
	for _, v := range allFriends {
		friendStatus, err := t.FriendGetConnectionStatus(v)
		if err != nil {
			continue
		}
		if friendStatus != gotox.TOX_CONNECTION_NONE {
			i++
		}
	}

	var (
		botname string
		displayUserInStatusText,
		addSonendofStatus bool
	)

	for _, t2 := range toxes {
		if t == t2.tox {
			botname = t2.botname
			displayUserInStatusText = t2.displayUserInStatusText
			addSonendofStatus = t2.addSonendofStatus
		}
	}

	middle := " "
	end := ""

	if displayUserInStatusText {
		if i == 1 {
			middle = " user in "
		} else {
			middle = " users in "
		}
	}
	if addSonendofStatus {
		if i > 1 {
			end = "s"
		}
	}

	t.SelfSetStatusMessage(strconv.Itoa(i) + middle + botname + end)
}

func updateTyping(t *gotox.Tox) {
	allFriends, err := t.SelfGetFriendlist()
	if err != nil {
		return
	}

	var lastFriendTyping uint32 = 4294967295
	var friendsTyping uint32

	for _, v := range allFriends {
		friendStatus, err := t.FriendGetConnectionStatus(v)
		if err != nil {
			continue
		}
		if friendStatus != gotox.TOX_CONNECTION_NONE {
			friendTyping, _ := t.FriendGetTyping(v)
			if friendTyping {
				friendsTyping++
				lastFriendTyping = v
			}
		}
	}

	for _, v := range allFriends {
		if friendsTyping >= 2 || (friendsTyping == 1 && lastFriendTyping != v) {
			t.SelfSetTyping(v, true)
		} else {
			t.SelfSetTyping(v, false)
		}
	}
}
