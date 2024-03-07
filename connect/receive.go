package connect

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/pkg/sftp"
)

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
		if fN == targetFile && !oSFs.IsDir() && Check_hash(fullpathClient, fullPath) {
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
	{
		_ , nameOfFile  := filepath.Split(Path_target)
		fmt.Printf("The file does not exist on your machine : %v\nDownloading %v ...\n", nameOfFile , humanize.Bytes(uint64(fileStat.Size())))
	}
	_, err = os.Stat(workdir_client + Path_target)
	if os.IsNotExist(err) {
		localfile, _ := os.Create(workdir_client + Path_target)
		_, err = io.Copy(localfile, openT)
		if err != nil {
			log.Println(err)
			return 
		}
		localfile.Close()
		} else {
			localfile, _ := os.OpenFile(workdir_client + Path_target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
			_, err = io.Copy(localfile, openT)
			if err != nil {
				log.Println(err)
				return 
			}
			localfile.Close()
		}
		defer openT.Close()
}