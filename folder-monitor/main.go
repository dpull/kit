package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("usage:folder-monitor dir")
		return
	}

	dir := os.Args[1]

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					log.Println("modified:", event.Name)
				}
				if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					log.Println("remove:", event.Name)
				}
				if event.Has(fsnotify.Create) {
					log.Println("create:", event.Name)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}
