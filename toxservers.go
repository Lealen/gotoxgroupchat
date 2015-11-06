package main

import (
	"encoding/hex"
	"fmt"

	"github.com/codedust/go-tox"
)

type Server struct {
	Address   string
	Port      uint16
	PublicKey []byte
}

func returnToxServer(Address string, Port uint16, PublicKey string) *Server {
	pubkey, err := hex.DecodeString(PublicKey)
	if err != nil {
		fmt.Println("[ERROR]", err)
	}

	return &Server{Address, Port, pubkey}
}

//# https://wiki.tox.chat/users/nodes
func connectToToxNetwork(tox *gotox.Tox) {
	var servers []*Server
	servers = append(servers, returnToxServer("144.76.60.215", 33445, "04119E835DF3E78BACF0F84235B300546AF8B936F035185E2A8E9E0A67C8924F"))   //sonOfRa
	servers = append(servers, returnToxServer("23.226.230.47", 33445, "A09162D68618E742FFBCA1C2C70385E6679604B2D80EA6E84AD0996A1AC8A074"))   //stal
	servers = append(servers, returnToxServer("178.21.112.187", 33445, "4B2C19E924972CB9B57732FB172F8A8604DE13EEDA2A6234E348983344B23057"))  //SylvieLorxu
	servers = append(servers, returnToxServer("195.154.119.113", 33445, "E398A69646B8CEACA9F0B84F553726C1C49270558C57DF5F3C368F05A7D71354")) //Munrek
	servers = append(servers, returnToxServer("192.210.149.121", 33445, "F404ABAA1C99A9D37D61AB54898F56793E1DEF8BD46B1038B9D822E8460FAB67")) //nurupo
	servers = append(servers, returnToxServer("46.38.239.179", 33445, "F5A1A38EFB6BD3C2C8AF8B10D85F0F89E931704D349F1D0720C3C4059AF2440A"))   //Martin Schröder
	servers = append(servers, returnToxServer("178.62.250.138", 33445, "788236D34978D1D5BD822F0A5BEBD2C53C64CC31CD3149350EE27D4D9A2F9B6B"))  //Impyy
	servers = append(servers, returnToxServer("130.133.110.14", 33445, "461FA3776EF0FA655F1A05477DF1B3B614F7D6B124F7DB1DD4FE3C08B03B640F"))  //Manolis
	servers = append(servers, returnToxServer("104.167.101.29", 33445, "5918AC3C06955962A75AD7DF4F80A5D7C34F7DB9E1498D2E0495DE35B3FE8A57"))  //noisykeyboard
	servers = append(servers, returnToxServer("205.185.116.116", 33445, "A179B09749AC826FF01F37A9613F6B57118AE014D4196A0E1105A98F93A54702")) //Busindre
	servers = append(servers, returnToxServer("198.98.51.198", 33445, "1D5A5F2F5D6233058BF0259B09622FB40B482E4FA0931EB8FD3AB8E7BF7DAF6F"))   //Busindre
	servers = append(servers, returnToxServer("80.232.246.79", 33445, "CF6CECA0A14A31717CC8501DA51BE27742B70746956E6676FF423A529F91ED5D"))   //fUNKIAM
	servers = append(servers, returnToxServer("108.61.165.198", 33445, "8E7D0B859922EF569298B4D261A8CCB5FEA14FB91ED412A7603A585A25698832"))  //ray65536
	servers = append(servers, returnToxServer("212.71.252.109", 33445, "04119E835DF3E78BACF0F84235B300546AF8B936F035185E2A8E9E0A67C8924F"))  //Kr9r0x
	servers = append(servers, returnToxServer("194.249.212.109", 33445, "3CEE1F054081E7A011234883BC4FC39F661A55B73637A5AC293DDF1251D9432B")) //fluke571
	servers = append(servers, returnToxServer("185.25.116.107", 33445, "DA4E4ED4B697F2E9B000EEFE3A34B554ACD3F45F5C96EAEA2516DD7FF9AF7B43"))  //MAH69K
	servers = append(servers, returnToxServer("192.99.168.140", 33445, "6A4D0607A296838434A6A7DDF99F50EF9D60A2C510BBF31FE538A25CB6B4652F"))  //WIeschie

	for _, server := range servers {
		err := tox.Bootstrap(server.Address, server.Port, server.PublicKey)
		if err != nil {
			fmt.Println("[ERROR]", err)
		}
	}
}
