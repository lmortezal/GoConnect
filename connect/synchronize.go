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
func synchronizeFiles(fullPath, workdir_target string) bool {
	var targetFile string = strings.Split(fullPath, "/")[len(strings.Split(fullPath, "/"))-1:][0]

	list_Dir := lsFiles_client()
	for _, fullpathClient := range list_Dir {
		_, fN := filepath.Split(fullpathClient)
		oSFs, _ := os.Stat(fullpathClient)
		if fN == targetFile && !oSFs.IsDir() && Check_hash(fullpathClient, fullPath) {
			fmt.Printf("%v already exists on the server.\nsize: %v\n", fN, humanize.Bytes(uint64(oSFs.Size())))
			return false
		}
	}
	sftpDownloader(fullPath, strings.Replace(fullPath, workdir_target, "", 1))
	return true
}


func IsExist(path string) bool {
	sftpSession, err := sftp.NewClient(client)
	if err != nil {
		log.Println(err)
	}
	defer sftpSession.Close()
	_, err = sftpSession.Stat(path)
	if err != nil {
		log.Println(err)
		if os.IsNotExist(err) {
			return false
		}
	}
	return true

}


// Get list of the files in the directory (client)
func lsFiles_client() []string{
	var list_Dir []string
	filepath.Walk(workdir_client, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() || info.Name()[0] == '.' {
			list_Dir = append(list_Dir, path)
		}
		return nil
	})
	return list_Dir
}


// Get list of the files in the directory (server)
// plus create the directory if it does not exist
func lsFiles_server(workdir string) (files []string) {
	_, DirName := filepath.Split(workdir_client)
	_, err := os.Stat(workdir_client)
	if err != nil {
		log.Printf("I cant create %v Dir or the Dir is already exist : %v\n", DirName, err)
		os.Mkdir(workdir_client, os.ModePerm)
	}
	GoDir, _ := os.Stat(workdir_client)
	if !GoDir.IsDir() {
		log.Printf("%v is not a directory\nPlz remove it and run again", workdir_client)
		return
	}
	sftpSessions, err := sftp.NewClient(client)
	if err != nil {
		log.Println(err)
		return
	}
	defer sftpSessions.Close()
	walkFile := sftpSessions.Walk(workdir)
	for walkFile.Step() {
		if !walkFile.Stat().IsDir() {
			files = append(files, walkFile.Path())
		} else {
			if _, err := os.Stat(strings.Replace(walkFile.Path(), workdir, "", 1)); os.IsNotExist(err) {
				os.Mkdir(workdir_client+strings.Replace(walkFile.Path(), workdir, "", 1), os.ModePerm)
			}
		}
	}
	return files
}

// Download the file from the server
func sftpDownloader(fullPath_target, Path_target string)  {
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
	fileStat, err := openT.Stat()
	if err != nil {
		log.Println(err)
	}
	if fileStat.IsDir() {
		return
	}

	{
		_, nameOfFile := filepath.Split(Path_target)
		fmt.Printf("The file does not exist on your machine : %v\nDownloading %v ...\n", nameOfFile, humanize.Bytes(uint64(fileStat.Size())))
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
		localfile, _ := os.OpenFile(workdir_client+Path_target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
		_, err = io.Copy(localfile, openT)
		if err != nil {
			log.Println(err)
			return
		}
		localfile.Close()
	}
}

func sftpUploader(file string , workdir_target string , filedownloaded []string ){
	// list_dir_target := lsFiles_server(workdir_target)
	sftpSesstion , err := sftp.NewClient(client)
	if err != nil {
		log.Println(err)
	}
	defer sftpSesstion.Close()
	OpenC , err := os.Open(file)
	if err != nil {
		log.Println(err)
	}
	defer OpenC.Close()
	fileStat , _ := OpenC.Stat()
	if fileStat.IsDir() {
		return
	}
	for _, fileDownloaded := range filedownloaded {
		if file == fileDownloaded {
			return
		}
	}
	_, err = sftpSesstion.Stat(workdir_target + filepath.SplitList(file)[len(filepath.SplitList(file)) - 1])	
	if err != nil {
		// log.Println(err)
		// log.Println("The file does not exist on the server")
		// log.Println("Uploading ...")
		var remoteFile *sftp.File
		remoteFile , err = sftpSesstion.Create(workdir_target + filepath.SplitList(file)[len(filepath.SplitList(file)) - 1])
		if err != nil {
			log.Println(err)
		}
		_, err = io.Copy(remoteFile , OpenC)
		if err != nil {
			log.Println(err)
		}
		remoteFile.Close()
	}


}
