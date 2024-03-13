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

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
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
	fmt.Printf("Enter your password:\n")
	password , err := term.ReadPassword(int(syscall.Stdin))
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
	var HostKeyCallB ssh.HostKeyCallback

	knownhostAddress, _ := PathFixer("~/.ssh/known_hosts")
	HostKeyCallB , err := knownhosts.New(knownhostAddress)
	if err != nil {
		log.Println(err)
		
	}

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
			// HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			HostKeyCallback: HostKeyCallB,
		}
		return config
	}
	config := &ssh.ClientConfig{
		User: user_ssh,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(singer),
		},
		HostKeyCallback: HostKeyCallB, // Only use this if you're sure about the server's identity
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
	workdir_client = filepath.Join(workdir_client , "/" , "GoConnect_" + ip_ssh )
	sshConnect(ip_ssh + ":" + port)
	var fileDownloaded = make([]string,0)


	files_server := lsFiles_server(workdir_target , true)
	for _, file := range files_server {
		if file == "" {
			continue
		}
		if download(file, workdir_target){
			fileDownloaded = append(fileDownloaded, file)
		}
	}
	files_client := lsFiles_client(workdir_target , true)
	for _ , file := range files_client{
		if file ==  ""{
			continue
		}
		// TODO check downloaded file
		upload(file , workdir_target)

	}

	// files_client := lsFiles_client()
	// for _, file := range files_client {
	// 	if file == "" {
	// 		continue
	// 	}
	// 	sftpUploader(file, workdir_target,fileDownloaded)
	// }


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
