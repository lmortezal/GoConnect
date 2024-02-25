package main

import (
	"fmt"
	"os"

	"github.com/lmortezal/GoConnect/connect"
)



func main() {
	a := os.Args
	if len(a) == 1{
		fmt.Println(`this is a Syncer app
	tempalate for a syncer app is as follows:
	  ./Goconnect <workdir>(default is ./) <destination>(e.g. User@ServerIp:/path/you/want/sync) <port>(default is 22) `)
		return
	}
	connect.Connect(a[1] , a[2] )

}



