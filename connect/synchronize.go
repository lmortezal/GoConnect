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

func upload(pathClient, pathServer string) bool {
	lsfilesServer := lsFiles_server(pathServer , false)
	sftpsession , _ := sftp.NewClient(client)
	defer sftpsession.Close()
	_ , fn := filepath.Split(pathClient)
	for _ , file := range lsfilesServer{
		openT,_ := sftpsession.Stat(file)	
		_ , targetFile := filepath.Split(file)	
		oSFs, _ := os.Stat(pathClient)
		if  fn == targetFile && !openT.IsDir() && Check_hash(pathClient , file) {
			fmt.Printf("%v already exists on the server.\nsize: %v\n", fn, humanize.Bytes(uint64(oSFs.Size())))
			return false
		}

	}
	sftpUploader(pathClient , strings.Replace(pathClient, workdir_client, "", 1))
	return true
}

// Get the file from the server with Check_hash
func download(fullPath, workdir_target string) bool {
	// var targetFile string = strings.Split(fullPath, "/")[len(strings.Split(fullPath, "/"))-1:][0]
	_ , targetFile := filepath.Split(fullPath)
	list_Dir := lsFiles_client("" , false)
	for _ , fullpathClient := range list_Dir {
		_ , fN := filepath.Split(fullpathClient)
		oSFs, _ := os.Stat(fullpathClient)
		if fN == targetFile && !oSFs.IsDir() && Check_hash(fullpathClient, fullPath) {
			fmt.Printf("%v already exists on the client.\nsize: %v\n", fN, humanize.Bytes(uint64(oSFs.Size())))
			return false
		}
	}
	sftpDownloader(fullPath, strings.Replace(fullPath, workdir_target, "", 1))
	return true
}





// Get list of the files name in the directory (client)
func lsFiles_client(workdir any , Cdir bool) (files []string){
	var workdirS string
	switch workdir.(type) {
		case string:
			workdirS = workdir.(string)
		case nil:
			return nil
		default:
			return nil
	}
	sftpsession , _ := sftp.NewClient(client)
	_ = sftpsession
	filepath.Walk(workdir_client, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() || info.Name()[0] == '.' {
			files = append(files, path)
		}else {
			if Cdir{
				sftpsession.Mkdir(filepath.ToSlash(filepath.Join(workdirS , strings.Replace(path, workdir_client, "", 1))))
				// fmt.Println(filepath.ToSlash(filepath.Join(workdirS , strings.Replace(path, workdir_client, "", 1))))
				// FIX PATH
			}
		}
		return nil
	})
	return files
}


// Get list of the files name in the directory (server)
// plus create the directory if it does not exist
func lsFiles_server(workdir string , Cdir bool) (files []string) {
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
			if _, err := os.Stat(strings.Replace(walkFile.Path(), workdir, "", 1)); os.IsNotExist(err)  && Cdir {
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

func sftpUploader(fullpathTarget string , pathTarget string  ){
	// list_dir_target := lsFiles_server(workdir_target)
	sftpSesstion , err := sftp.NewClient(client)
	if err != nil {
		log.Println(err)
	}
	defer sftpSesstion.Close()
	OpenC , err := os.Open(fullpathTarget)
	if err != nil {
		log.Println(err)
	}
	defer OpenC.Close()
	fileStat , _ := OpenC.Stat()
	if fileStat.IsDir() {
		return
	}
	{
		_, nameOfFile := filepath.Split(fullpathTarget)
		fmt.Printf("The file does not exist on your server : %v\nUploading %v ...\n", nameOfFile, humanize.Bytes(uint64(fileStat.Size())))
	}
	// FIX ToSlash
	_ , err = sftpSesstion.Stat(filepath.ToSlash(filepath.Join(workdir_target , pathTarget)))
	if	err != nil{
		sftpSesstion.Create(filepath.ToSlash(filepath.Join(workdir_target , pathTarget)))
	}
	serverfile , _ := sftpSesstion.OpenFile(filepath.ToSlash(filepath.Join(workdir_target , pathTarget)) , os.O_RDWR|os.O_CREATE|os.O_TRUNC)
	_ , err = io.Copy(serverfile , OpenC)
	if err != nil {
		log.Println(err)
		return
	}
	defer serverfile.Close()
	
}
