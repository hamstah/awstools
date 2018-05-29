package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/fsnotify/fsnotify"
)

func loadConfig(filename string) (*Config, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	var config Config
	err = json.Unmarshal(raw, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func updateConfig(filename string) {
	newConfig, err := loadConfig(filename)
	if err != nil {
		log.Println("Failed to load new config", err)
	} else {
		configM.Lock()
		config = newConfig
		configM.Unlock()
	}
}

func watchConfig(filename string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ugh")
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					updateConfig(filename)
					log.Println("modified file:", event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(filename)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
