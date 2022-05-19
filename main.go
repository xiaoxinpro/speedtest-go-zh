package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	_ "time/tzdata"

	"github.com/xiaoxinpro/speedtest-go-zh/config"
	"github.com/xiaoxinpro/speedtest-go-zh/database"
	"github.com/xiaoxinpro/speedtest-go-zh/results"
	"github.com/xiaoxinpro/speedtest-go-zh/web"

	_ "github.com/breml/rootcerts"
	log "github.com/sirupsen/logrus"
)

var (
	optConfig = flag.String("c", "./config/settings.toml", "config file to be used, defaults to settings.toml in the same directory")
)

func main() {
	InitFile()
	flag.Parse()
	conf := config.Load(*optConfig)
	web.SetServerLocation(&conf)
	results.Initialize(&conf)
	database.SetDBInfo(&conf)
	log.Fatal(web.ListenAndServe(&conf))
}

//InitFile 初始化文件
func InitFile() {
	if isOK,_ := PathExists("./config"); isOK == false {
		os.MkdirAll("./config", os.ModePerm)
	}
	if isOK,_ := PathExists("./config/settings.toml"); isOK == false {
		CopyFile("./settings.toml", "./config/settings.toml")
		os.Chmod("./config/settings.toml", os.ModePerm)
	}
}

//PathExists 判断一个文件或文件夹是否存在
func PathExists(path string) (bool,error) {
	_,err := os.Stat(path)
	if err == nil {
		return true,nil
	}
	if os.IsNotExist(err) {
		return false,nil
	}
	return false,err
}

//CopyFile 复制文件
func CopyFile(srcFileName string, dstFileName string) (written int64, err error) {
	srcFile, err := os.Open(srcFileName)
	if err != nil {
		fmt.Printf("open file err = %v\n", err)
		return
	}
	defer srcFile.Close()

	//open dstFileName
	dstFile, err := os.OpenFile(dstFileName, os.O_WRONLY | os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("open file err = %v\n", err)
		return
	}
	defer dstFile.Close()
	return io.Copy(dstFile, srcFile)
}