package main

import (
	"bytes"
	"fmt"
	"log"
	"golang.org/x/crypto/ssh"
	"github.com/pkg/sftp"
)



func main() {
	// var hostKey ssh.PublicKey
	
	// An SSH client is represented with a ClientConn.
	//
	// To authenticate with the remote server you must pass at least one
	// implementation of AuthMethod via the Auth field in ClientConfig,
	// and provide a HostKeyCallback.
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("123"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", "---:22", config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer client.Close()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("cd / && ls"); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())
	var sftpClient *sftp.Client
	sftpClient, err = sftp.NewClient(client)
	if err != nil {
		log.Fatal(err)
	}
	defer sftpClient.Close()
	// creata a txt file in the remote server
	f , err := sftpClient.Create("/home/ubuntu/test.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(f)
	if err := f.Chmod(0777); err != nil {
		log.Fatal(err)
	}
	f.Close()
	
	w := sftpClient.Walk("/home")
	for w.Step() {
		if w.Err() != nil {
			continue
		}
		fmt.Println(w.Path())
	}
	
}


