package connect

import (
	"bytes"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var (
	privateKeyPath string = "/home/morteza/.ssh/id_rsa"
	Password       string
	ip_ssh         string
	user_ssh       string
	workdir_client string
	workdir_target *string
	session        *ssh.Session
	// protect the password and save it in a file
)

func loger() {
	_, _, line, _ := runtime.Caller(1)
	log.Printf("--%v--\n", line)
}

func error_check(err error, line uint) {
	// each error shoud be stop the program
	if err != nil {
		log.Println(err, "in line: ", line)
	}
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
	if privateKeyPath == "" {
		config := &ssh.ClientConfig{
			User: user_ssh,
			Auth: []ssh.AuthMethod{
				ssh.Password("123"),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		return config
	} else if Password == "" {
		config := &ssh.ClientConfig{
			User: user_ssh,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(singer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Only use this if you're sure about the server's identity
		}
		return config
	}
	return nil
}
func Check_Dir(path fs.DirEntry) bool {
	// check the directory is it dir or not
	return path.IsDir()
}


func Chech_hash(file_client, file_target string) bool {

	// check the hash of the file
	// from server check the hash of the file
	md5_target, _ := session.Output("md5sum " + file_target)
	loger()
	fmt.Println(string(md5_target) + " target")
	md5_client, _ := exec.Command("md5sum", workdir_client+file_client).Output()
	loger()
	fmt.Println(strings.Split(string(md5_client)," ")[0] + " client------target " + string(md5_target))
	return strings.Split(string(md5_client)," ")[0] == string(md5_target)
}

func GetHash(){
	//first connect to the server and get the hash of the file
	config := Check_method_connect()
	fmt.Println(ip_ssh)
	client, err := ssh.Dial("tcp", ip_ssh+":22", config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer client.Close()
	session, err = client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("cd " + *workdir_target + " && ls -ltrh"); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	var _ = strings.Split(b.String(), "\n")
	fmt.Println(b.String())
}


func GetFile(client *ssh.Client, fileName string) {
	var sftpClient *sftp.Client
	sftpClient, err := sftp.NewClient(client)
	defer sftpClient.Close()

	exePath, err := os.Executable()
	if err != nil {
		log.Println(err)
		fmt.Println(exePath)
	}
	// os.Chdir(exePath)
	os.Chdir("/home/morteza/Desktop/GoConnect/")
	workdir_client = "/home/morteza/Desktop/GoConnect/"
	list_Dir, _ := os.ReadDir(".")
	for _, Name_of_File := range list_Dir {
		if Name_of_File.Name() == fileName && Name_of_File.IsDir() && Chech_hash(Name_of_File.Name(), fileName) {
			fmt.Println("The file is already exist")
		}
		fmt.Printf(`Name_of_File.Name() : %v
fileName : %v
!Check_Dir(Name_of_File) : %v
Chech_hash(Name_of_File.Name() : %v
` , Name_of_File.Name() , fileName , !Check_Dir(Name_of_File) , Chech_hash(Name_of_File.Name(), fileName)) 
	}
	os.Exit(1)
	fmt.Println("The file is not exist")

}

func Connect(sshAddr string, address_from_main string) {
	workdir_target = &address_from_main
	if sshAddr == "" || len(strings.Split(sshAddr, "@")) < 2 {
		log.Println(`Wrong input!!`)
		return
	}
	user_ssh = strings.Split(sshAddr, "@")[0]
	ip_ssh = strings.Split(strings.Split(sshAddr, "@")[1], ":")[0]
	GetHash()

	config := Check_method_connect()
	client, err := ssh.Dial("tcp", ip_ssh+":22", config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer client.Close()
	log.Println("Connected to the server")

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err = client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()
	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("cd " + *workdir_target + " && ls"); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	var list = strings.Split(b.String(), "\n")
	// fmt.Println(b.String())
	for i := 0; i < len(list); i++ {
		if list[i] == "" {
			continue
		}
		GetFile(client, list[i])
		fmt.Println("----------------" + list[i] + "----------------")
		fmt.Println(list[i])
	}

	var sftpClient *sftp.Client
	sftpClient, err = sftp.NewClient(client)
	if err != nil {
		log.Fatal(err)
	}
	defer sftpClient.Close()
	// creata a txt file in the remote server
	f, err := sftpClient.Create("/home/ubuntu/test.txt")
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(f)
	if err := f.Chmod(0777); err != nil {
		log.Fatal(err)
	}
	f.Close()

	w := sftpClient.Walk("/home/ubuntu/")
	for w.Step() {
		if w.Err() != nil {
			continue
		}
		// fmt.Println(w.Path())
	}
	// fi , err := sftpClient.Lstat("/home/ubuntu/test.txt")
	if err != nil {
		log.Fatal(err)
	}
	// log.Println(fi)

}
