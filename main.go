package main

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/lmortezal/GoConnect/connect"
)

func main() {
	var (
		destination string
		source string
		port uint
	)
	
	dir , _ := os.Getwd()
	flag.StringVar(&destination, "d", dir , "destination directory (e.g. /home/user/GoConnect )")
	flag.StringVar(&source, "s", "" , "source directory (e.g. user@golang:/home/target/GoConnect )")
	flag.UintVar(&port, "p", 22 , "port number")
	flag.Parse()
	flag.Func("help", "show help" , func(s string) error { flag.PrintDefaults(); return nil })
	if len(os.Args) == 1{
		flag.PrintDefaults()
		return
	}else if source == "" {
		flag.PrintDefaults()
		return
	}
	var sources [3]string
	sources[0] = strings.Split(source , "@")[0]
	sources[2] = strings.Split(source , ":")[1] 
	sources[1] = strings.Split(strings.Split(source, "@")[1],":")[0]

	connect.Connect(destination , strconv.Itoa(int(port)) , sources)


}




