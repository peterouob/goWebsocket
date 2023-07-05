package main

import (
	"context"
	"log"
	"net/http"
)

func main() {
	setupApi()
	log.Fatalln(http.ListenAndServe(":8082", nil))
}

func setupApi() {
	ctx := context.Background()
	manger := NewManager(ctx)
	http.Handle("/", http.FileServer(http.Dir("./front"))) //加載前端
	http.HandleFunc("/ws", manger.serveWs)
	http.HandleFunc("/login", manger.loginHandler)
}

//go run *.go 運行所有文件
