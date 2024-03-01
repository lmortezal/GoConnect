package connect

import (
	"crypto/md5"
	"fmt"
	"os"
	"io"

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