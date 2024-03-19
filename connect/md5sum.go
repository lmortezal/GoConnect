package connect

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"
)
func Md5sum(FilePath string) string{
	file , err := os.Open(FilePath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	hash := md5.New()
	_, err = io.Copy(hash, file)

	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}


// Get the hash of the file
func Check_hash(FullPath_client, fullPath_target string) bool {
	fmt.Println(fullPath_target)
	fmt.Println("---------------")
	fmt.Println(FullPath_client)
	
	// check the hash of the file
	// from server check the hash of the file
	md5_target := Run_Command("md5sum " + fullPath_target)
	md5_target = strings.Split(string(md5_target), " ")[0]

	// md5_client_, err := exec.Command("md5sum" , filepath.Join(workdir_client +"/"+ file_client)).Output()
	// md5_client := strings.Split(string(md5_client_), " ")[0][1:]

	md5_client := Md5sum(FullPath_client)
	return md5_target == md5_client
}