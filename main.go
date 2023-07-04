package main

import (
	"log"
	"net/http"
)

func main() {
	setupApi()
	log.Fatalln(http.ListenAndServe(":8082", nil))
}

func setupApi() {
	manger := Manager{}
	http.Handle("/", http.FileServer(http.Dir("./front"))) //加載前端
	http.HandleFunc("/ws", manger.serveWs)
}

//go run *.go 運行所有文件
