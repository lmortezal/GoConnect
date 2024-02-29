package connect

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var (
	privateKeyPath string
	ip_ssh         string
	user_ssh       string
	workdir_client string
	workdir_target string
	client         *ssh.Client
	err            error
	// protect the password and save it in a file
)

func GivePassword() string {
	LABLE:
	fmt.Println("Enter your password: ")
	password , err := term.ReadPassword(0)
	if err != nil {
		log.Println(err)
		goto LABLE
	}
	return string(password)
}

func PathFixer(path string) (string, error) {
	// fix the directory Path
	if !strings.HasPrefix(path, "~") {
		return path, nil
	} else if strings.HasPrefix(path, ".") {
		// TODO : test this part
		return filepath.Abs(path)
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	// Replace ~ with the user's home directory
	return filepath.Join(usr.HomeDir, path[1:]), nil
}

func loger() {
	_, _, line, _ := runtime.Caller(1)
	log.Printf("--%v--\n", line)
}

func Check_method_connect() *ssh.ClientConfig {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Println(err)
	}
	singer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Println(err)
	}
	if singer == nil {
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

func Chech_hash(file_client, file_target string) bool {
	// check the hash of the file
	// from server check the hash of the file
	md5_target := Run_Command("md5sum " + workdir_target + "/" + file_target)
	md5_client, _ := exec.Command("md5sum", workdir_client+"/"+file_client).Output()
	return strings.Split(string(md5_client), " ")[0] == strings.Split(string(md5_target), " ")[0]
}

func GetFile(fileName_target string) {
	list_Dir, _ := os.ReadDir(".")
	for _, Name_of_File := range list_Dir {
		if Name_of_File.Name() == fileName_target && !Name_of_File.IsDir() && Chech_hash(Name_of_File.Name(), fileName_target) {
			fmt.Println("The file is already exist and same as in server \nfile name : \n" + Name_of_File.Name())
			return
		}
	}

	fmt.Println("The file is not exist: \nfile name : \n" + fileName_target)
	fmt.Printf("Downloading ....\n")
	sftpDownloader(fileName_target)
}

func sftpDownloader(fileName_target string) bool {
	// download the file from the server
	sftpSession, err := sftp.NewClient(client)
	if err != nil {
		log.Println(err)
	}
	defer sftpSession.Close()
	var openT *sftp.File
	openT, err = sftpSession.Open(workdir_target + "/" + fileName_target)
	if err != nil {
		log.Println(err)
	}
	defer openT.Close()
	_, err = os.Stat(fileName_target)
	if os.IsNotExist(err) {
		localfile, _ := os.Create(fileName_target)
		_, err = io.Copy(localfile, openT)
		if err != nil {
			log.Println(err)
			return false
		}
		localfile.Close()
	} else {
		localfile, _ := os.OpenFile(fileName_target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
		_, err = io.Copy(localfile, openT)
		if err != nil {
			log.Println(err)
			return false
		}
		localfile.Close()
	}
	defer openT.Close()
	return true
}

func Run_Command(command string) (resault string) {
	//first connect to the server and get the hash of the file
	fmt.Println("command : " + command)
	session_Scope, err := client.NewSession()
	if err != nil {
		log.Println("Failed to dial: ", err)
	}
	loger()
	defer session_Scope.Close()
	var b bytes.Buffer
	session_Scope.Stdout = &b
	if err := session_Scope.Run(command); err != nil {
		log.Println("Failed to dial: ", err)

	}
	return b.String()
}

func Initailize(addr string) {
	loger()
	privateKeyPath, err = PathFixer("~" + "/.ssh/id_rsa2")
	config := Check_method_connect()
	fmt.Println(addr)
	client, err = ssh.Dial("tcp", addr, config)
	// defer client.Close()
	loger()
	if err != nil {
		log.Println(err)
	}

	os.Chdir(workdir_client)

}

func Connect(destination string, port string, sources [3]string) {
	ip_ssh = sources[1]
	user_ssh = sources[0]
	workdir_target = sources[2]
	workdir_client = destination
	// print all variable
	fmt.Println(ip_ssh)
	fmt.Println(user_ssh)
	fmt.Println(workdir_target)
	fmt.Println(workdir_client)
	Initailize(ip_ssh + ":" + port)

	list := strings.Split((Run_Command("cd " + workdir_target + " && ls -A")), "\n")
	for _, res := range list {
		if res == "" {
			continue
		}
		loger()
		GetFile(res)
	}
	defer client.Close()
	fmt.Println("Finished")
}
