package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nvgruel123/cs686-blockchain-p3-nvgruel123/p3"
)

func main() {
	router := p3.NewRouter()
	if len(os.Args) == 4 {
		p3.SetSelfPort(os.Args[1])
		p3.SetPrivateKey(os.Args[2], os.Args[3])
		log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	} else if len(os.Args) == 2 {
		p3.SetSelfPort(os.Args[1])
		log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	} else {
		log.Fatal(http.ListenAndServe(":6686", router))
	}
}
