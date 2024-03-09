package connect

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	// "time"
)

var (
	ip_ssh         string
	user_ssh       string
	workdir_client string
	client         *ssh.Client
	err            error
	// protect the password and save it in a file
)

// Get the password from the user
func GivePassword() string {
	GETPASSAGAIN:
	fmt.Printf("Enter your password:  ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Printf("Error to read password\nError:  %v\n", err)
		goto GETPASSAGAIN
	} else if len(password) == 0 {
		fmt.Printf("Password can not be empty\n")
		goto GETPASSAGAIN
	}
	return string(password)
}

// check the path 
func PathFixer(path string) (string, error) {
	// fix the directory Path

	if strings.HasPrefix(path, ".") {
		return filepath.Abs(path)
	} else if !strings.HasPrefix(path, "~") {
		return path, nil
		// TODO : test this part
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	// Replace ~ with the user's home directory
	return filepath.Join(usr.HomeDir, path[1:]), nil
}

// print the line of the code
func loger() {
	_, _, line, _ := runtime.Caller(1)
	log.Printf("--%v--\n", line)
}

// Get the private key from the ssh directory
func GetPrivateKey() string {
	var keyArray []string = []string{"id_rsa", "id_ecdsa", "id_ecdsa_sk", "id_ed25519", "id_ed25519_sk", "id_dsa", "id_xmss"}
	for _, key := range keyArray {
		FullAddress, _ := PathFixer("~/.ssh/" + key)
		if _, err := os.ReadFile(FullAddress); err == nil {
			fmt.Printf("Use the key : %v\n", key)
			return "~/.ssh/" + key
		}
	}
	return ""
}

// Get list of the files in the directory (server)
// plus create the directory if it does not exist
func lsFiles(workdir string) (files []string){
	sftpSessions , err := sftp.NewClient(client)
	if err != nil{
		log.Println(err)
		return 
	}
	defer sftpSessions.Close()
	walkFile := sftpSessions.Walk(workdir)
	for walkFile.Step(){
		if !walkFile.Stat().IsDir(){
			files = append(files, walkFile.Path())
		}else {
			if _, err := os.Stat(strings.Replace(walkFile.Path(),workdir , "", 1)); os.IsNotExist(err) {
				os.Mkdir(workdir_client + strings.Replace(walkFile.Path(),workdir , "", 1) , os.ModePerm)
			}
		}
	}
	return files
}

// Run the command on the server
func Run_Command(command string) (resault string) {
	//first connect to the server and get the hash of the file
	session_Scope, err := client.NewSession()
	if err != nil {
		log.Println("Failed to dial: ", err)
	}
	defer session_Scope.Close()
	var b bytes.Buffer
	session_Scope.Stdout = &b
	if err := session_Scope.Run(command); err != nil {
		log.Println("Failed to dial: ", err)
		
	}
	
	return b.String()
}

// Check the authentication method and return the config
func Check_method_connect(Pkey *bool) *ssh.ClientConfig {
	// TODO : fix authentication check
	Paddress, _ := PathFixer(GetPrivateKey())
	key, err := os.ReadFile(Paddress)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
	}
	singer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Printf("Failed to parse private key: %v", err)
	}
	if singer == nil || !*Pkey {
		Password := GivePassword()
		config := &ssh.ClientConfig{
			User: user_ssh,
			Auth: []ssh.AuthMethod{
				ssh.Password(Password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		return config
	}
	config := &ssh.ClientConfig{
		User: user_ssh,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(singer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Only use this if you're sure about the server's identity
	}
	return config
}

// Initialize the connection and try to connect to the server
func sshConnect(addr string) {
	var Pkey bool = true
	GETPASS:
	config := Check_method_connect(&Pkey)
	client, err = ssh.Dial("tcp", addr, config)
	// defer client.Close()
	if err != nil {
		log.Println(err)
		Pkey = false
		goto GETPASS
	}
	os.Chdir(workdir_client)

}


// Start GoConnect from here
// source : user@ip:port
// destination : the path of the directory
// port : the port of the ssh
func Initailize(destination string, port string, sources [3]string) {
	ip_ssh = sources[1]
	user_ssh = sources[0]
	workdir_target := sources[2]
	workdir_client, _ = filepath.Abs(destination)
	sshConnect(ip_ssh + ":" + port)
	files := lsFiles(workdir_target)
	for _, file := range files {
		if file == "" {
			continue
		}
		synchronizeFiles(file, workdir_target)
	}
	
	// for {
	// 	files := lsFiles(workdir_target)
	// 	for _, file := range files {
	// 		if file == "" {
	// 			continue
	// 		}
	// 		GetFile(file, workdir_target)
	// 	}
	// 	time.Sleep(5 * time.Second)
	// }

	defer client.Close()
	fmt.Println("Finished")
}
