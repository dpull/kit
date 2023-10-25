package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dpull/kit/coffer/filesystem"
	"golang.org/x/net/webdav"
)

type config struct {
	HttpAddr string            `json:"http_addr"`
	Folder   string            `json:"folder"`
	FSType   string            `json:"fs_type"`
	FSParam  map[string]string `json:"fs_param"`
}

func initLog() (*os.File, error) {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	fd, err := os.OpenFile(*logPath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	log.SetOutput(fd)
	return fd, nil
}

func readConfig(conf *config) error {
	data, err := ioutil.ReadFile(*configPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, conf)
}

var configPath = flag.String("c", "config.json", "config file")
var logPath = flag.String("l", "coffer.log", "log file")

func main() {
	flag.Parse()
	// startHttpProfile(8080)

	fd, err := initLog()
	if err != nil {
		log.Fatalf("init log failed:%+v", err)
	}
	defer fd.Close()

	var conf config
	err = readConfig(&conf)
	if err != nil {
		flag.Usage()
		log.Fatalf("read config failed:%=v", err)
	}

	fs, err := filesystem.Create(conf.FSType, conf.Folder, conf.FSParam)
	if err != nil {
		log.Fatalf("create file system failed: %+v", err)
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/", &webdav.Handler{
			FileSystem: fs,
			LockSystem: webdav.NewMemLS(),
		})
		err := http.ListenAndServe(conf.HttpAddr, mux)
		if err != nil {
			log.Fatalf("http serve failed: %+v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)
	<-ch
}
