package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/lmortezal/GoConnect/connect"
)

func main() {
	var (
		source string
		destination string
		port uint
	)
	dir , _ := os.Executable()
	flag.StringVar(&source, "s", dir , "source directory (e.g. /home/user/GoConnect )")
	flag.StringVar(&destination, "d", "" , "destination directory (e.g. user@golang:/home/target/GoConnect )")
	flag.UintVar(&port, "p", 22 , "port number")
	flag.Parse()
	flag.Func("help", "show help" , func(s string) error { flag.PrintDefaults(); return nil })
	if len(os.Args) == 1{
		flag.PrintDefaults()
		return
	}else if destination == "" {
		flag.PrintDefaults()
		return
	}
	var destinations [3]string
	destinations[0] = strings.Split(destination , "@")[0]
	destinations[2] = strings.Split(destination , ":")[1] 
	destinations[1] = strings.Split(strings.Split(destination, "@")[1],":")[0]
	fmt.Println(destinations)

	connect.Connect(source , port , destinations)


}



