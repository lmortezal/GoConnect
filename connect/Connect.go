package connect

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var (
	privateKeyPath string = "/home/morteza/.ssh/id_rsa"
	Password       string
	ip_ssh         string
	user_ssh       string
	workdir_client string
	workdir_target *string
	client         *ssh.Client
	err            error
	// protect the password and save it in a file
)

// TODO
func directoryPath() {
	// fix the directory Path
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
	md5_target := Run_Command("md5sum " + *workdir_target + "/" + file_target)
	md5_client, _ := exec.Command("md5sum", workdir_client+file_client).Output()
	return strings.Split(string(md5_client), " ")[0] == strings.Split(string(md5_target), " ")[0]
}

func GetFile(fileName_target string) {
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
		if Name_of_File.Name() == fileName_target && !Name_of_File.IsDir() && Chech_hash(Name_of_File.Name(), fileName_target) {
			fmt.Println("The file is already exist and same as in server \nfile name : \n" + Name_of_File.Name())
			return
		}
		// fmt.Printf(`Name_of_File.Name() : %v
		// fileName : %v
		// !Check_Dir(Name_of_File) : %v
		// Chech_hash(Name_of_File.Name() : %v
		// ` , Name_of_File.Name() , fileName , !Check_Dir(Name_of_File) , Chech_hash(Name_of_File.Name(), fileName))
	}

	fmt.Println("The file is not exist: \nfile name : \n" + fileName_target)
	fmt.Printf(`Downloading ....
plesae wait ....
`)
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
	openT, err = sftpSession.Open(*workdir_target + "/" + fileName_target)
	if err != nil {
		log.Println(err)
	}
	defer openT.Close()
	os.Chdir("/home/morteza/Desktop/GoConnect")
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
		localfile, _ := os.OpenFile(fileName_target , os.O_RDWR | os.O_CREATE | os.O_TRUNC , 0777)
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
	defer session_Scope.Close()
	var b bytes.Buffer
	session_Scope.Stdout = &b
	if err := session_Scope.Run(command); err != nil {
		log.Println("Failed to dial: ", err)

	}
	return b.String()
}

func Connect(sshAddr string, address_from_main string) {
	workdir_target = &address_from_main
	if sshAddr == "" || len(strings.Split(sshAddr, "@")) < 2 {
		log.Println(`Wrong input!!`)
		return
	}
	user_ssh = strings.Split(sshAddr, "@")[0]
	ip_ssh = strings.Split(strings.Split(sshAddr, "@")[1], ":")[0]

	// ssh connect scope
	config := Check_method_connect()
	fmt.Println(ip_ssh)
	client, _ = ssh.Dial("tcp", ip_ssh+":22", config)
	defer client.Close()
	loger()

	list := strings.Split((Run_Command("cd " + *workdir_target + " && ls -A")), "\n")
	for _, res := range list {
		if res == "" {
			continue
		}
		loger()
		GetFile(res)
	}
	// config := Check_method_connect()
	// client, err := ssh.Dial("tcp", ip_ssh+":22", config)
	// if err != nil {
	// 	log.Fatal("Failed to dial: ", err)
	// }
	// defer client.Close()
	// log.Println("Connected to the server")

	// // Each ClientConn can support multiple interactive sessions,
	// // represented by a Session.
	// session, err = client.NewSession()
	// if err != nil {
	// 	log.Fatal("Failed to create session: ", err)
	// }
	// defer session.Close()
	// // Once a Session is created, you can execute a single command on
	// // the remote side using the Run method.
	// var b bytes.Buffer
	// session.Stdout = &b
	// if err := session.Run("cd " + *workdir_target + " && ls"); err != nil {
	// 	log.Fatal("Failed to run: " + err.Error())
	// }
	// var list = strings.Split(b.String(), "\n")
	// // fmt.Println(b.String())
	// for i := 0; i < len(list); i++ {
	// 	if list[i] == "" {
	// 		continue
	// 	}
	// 	GetFile(client, list[i])
	// 	fmt.Println("----------------" + list[i] + "----------------")
	// 	fmt.Println(list[i])
	// }

	// var sftpClient *sftp.Client
	// sftpClient, err = sftp.NewClient(client)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer sftpClient.Close()
	// // creata a txt file in the remote server
	// f, err := sftpClient.Create("/home/ubuntu/test.txt")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // fmt.Println(f)
	// if err := f.Chmod(0777); err != nil {
	// 	log.Fatal(err)
	// }
	// f.Close()

	// w := sftpClient.Walk("/home/ubuntu/")
	// for w.Step() {
	// 	if w.Err() != nil {
	// 		continue
	// 	}
	// 	// fmt.Println(w.Path())
	// }
	// // fi , err := sftpClient.Lstat("/home/ubuntu/test.txt")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // log.Println(fi)

}
