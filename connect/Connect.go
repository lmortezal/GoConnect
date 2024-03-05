package connect

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"github.com/dustin/go-humanize"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
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
	fmt.Println("Enter your password: ")
	password, err := term.ReadPassword(0)
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

// Get the hash of the file
func Chech_hash(FullPath_client, fullPath_target string) bool {
	// check the hash of the file
	// from server check the hash of the file
	md5_target := Run_Command("md5sum " + fullPath_target)
	md5_target = strings.Split(string(md5_target), " ")[0]

	// md5_client_, err := exec.Command("md5sum" , filepath.Join(workdir_client +"/"+ file_client)).Output()
	// md5_client := strings.Split(string(md5_client_), " ")[0][1:]

	md5_client := Md5sum(FullPath_client)
	return md5_target == md5_client
}


// Get the file from the server with Check_hash
func GetFile(fullPath,workdir_target string) {
	var targetFile string = strings.Split(fullPath , "/")[len(strings.Split(fullPath , "/")) - 1:][0]
	
	var list_Dir2 []string
	filepath.Walk(workdir_client, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() || info.Name()[0] == '.' {
			list_Dir2 = append(list_Dir2, path)
		}
		return nil
	})

	// list_Dir, _ := os.ReadDir(".")
	for _ , fullpathClient := range list_Dir2 {
		_ , fN := filepath.Split(fullpathClient)
		oSFs , _ := os.Stat(fullpathClient)
		if fN == targetFile && !oSFs.IsDir() && Chech_hash(fullpathClient, fullPath) {
			fmt.Printf("%v already exists on the server.\nsize: %v\n",fN,  humanize.Bytes(uint64(oSFs.Size())))
			return
		}
	}
	sftpDownloader(fullPath , strings.Replace(fullPath , workdir_target , "", 1))
}
// Download the file from the server
func sftpDownloader(fullPath_target,Path_target string)  {
	// download the file from the server
	// var fullpathTarget string = strings.Split(fullPath_target , "/")[len(strings.Split(fullPath_target , "/")) - 1:][0]
	sftpSession, err := sftp.NewClient(client)
	if err != nil {
		log.Println(err)
	}
	defer sftpSession.Close()
	var openT *sftp.File
	openT, err = sftpSession.Open(fullPath_target)
	if err != nil {
		log.Println(err)
	}
	defer openT.Close()
	
	// if filename is directory
	fileStat , err:= openT.Stat()
	if err != nil {
		log.Println(err)
	}
	if fileStat.IsDir(){
		return 
	}

	fmt.Printf("The file does not exist on your machine : %v\nDownloading %v ...\n", Path_target , humanize.Bytes(uint64(fileStat.Size())))
	_, err = os.Stat(Path_target)
	if os.IsNotExist(err) {
		localfile, _ := os.Create(Path_target)
		_, err = io.Copy(localfile, openT)
		if err != nil {
			log.Println(err)
			return 
		}
		localfile.Close()
		} else {
			localfile, _ := os.OpenFile(Path_target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
			_, err = io.Copy(localfile, openT)
			if err != nil {
				log.Println(err)
				return 
			}
			localfile.Close()
		}
		defer openT.Close()
}

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
				os.Mkdir(strings.Replace(walkFile.Path(),workdir , "", 1) , os.ModePerm)
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
func Initailize(addr string) {
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
func Connect(destination string, port string, sources [3]string) {
	ip_ssh = sources[1]
	user_ssh = sources[0]
	workdir_target := sources[2]
	workdir_client = destination

	Initailize(ip_ssh + ":" + port)
	files := lsFiles(workdir_target)
	fmt.Println(files)
	for _, file := range files {
		if file == "" {
			continue
		}
		GetFile(file, workdir_target)
	}
	defer client.Close()
	fmt.Println("Finished")
}
