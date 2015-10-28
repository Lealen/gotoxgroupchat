package main

import (
	"encoding/hex"

	"github.com/AllenDang/simhash"
	"github.com/codedust/go-tox"
)

func getFriendName(t *gotox.Tox, friendNumber uint32) string {
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

	addtoname := " (" + string('ア'+friendNumber*2) + " " + string(pubstr) + "...)"

	return name + addtoname
}

func searchForUser(t *gotox.Tox, searchfor string) (friendNumber uint32) {
	allFriends, err := t.SelfGetFriendlist()
	if err != nil {
		return
	}

	friendNumber = 4294967295
	var likeness float64

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

		tmpLikeness := simhash.GetLikenessValue(searchfor, name)
		if tmpLikeness > likeness {
			friendNumber = v
			likeness = tmpLikeness
		}

		tmpLikeness = simhash.GetLikenessValue(searchfor, string(pubstr))
		if tmpLikeness > likeness {
			friendNumber = v
			likeness = tmpLikeness
		}
	}

	return
}
