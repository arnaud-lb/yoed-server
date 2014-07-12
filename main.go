package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"github.com/gorilla/mux"
)

type yoedConfig struct {
	Listen   string `json:"listen"`
}

type yoedHandler interface {
	Handle(username string)
}

func loadConfig(configPath string) (*yoedConfig, error) {

	configFile, err := os.Open(configPath)

	if err != nil {
		return nil, err
	}

	configJson, err := ioutil.ReadAll(configFile)

	if err != nil {
		return nil, err
	}

	config := &yoedConfig{}

	if err := json.Unmarshal(configJson, config); err != nil {
		return nil, err
	}

	return config, nil
}

func main() {

	config, err := loadConfig("./config.json")

	if err != nil {
		panic(fmt.Sprintf("failed loading config: %s", err))
	}

	handlers := make(map[string]bool, 0)

	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		callbackUrl := r.FormValue("callback_url")
		log.Printf("subscribe %s", callbackUrl)
		handlers[callbackUrl] = true
	})
	router.HandleFunc(`/yoed/{handle:[a-z0-9]+}`, func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		handle := vars["handle"]
		username := r.FormValue("username")
		log.Printf("got a YO from %s on %s", username, handle)

		if 0 == len(handlers) {
			log.Printf("No handler registered")
		} else {
			for handler, _ := range handlers {
				log.Printf("Dispatch to handler %s", handler)
				resp, err := http.PostForm(handler, url.Values{"username":{username}})
				
				if err != nil {
					log.Printf("Error while dispatching message to %s: %s", handler, err)
					log.Printf("Remove handler %s", handler)
					delete(handlers, handler)
				} else {
					log.Printf("Handler %s status: %s", handler, resp.Status)
				}
			}
		}
	})

	server := http.Server{
		Addr:    config.Listen,
		Handler: router,
	}

	log.Printf("Listening...")

	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}

}