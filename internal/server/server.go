package server

import (
	"fmt"
	"net/http"
)

func NewServer() {

	http.Handle("/", http.FileServer(http.Dir("C://Practicum//sprint-final//web")))
	fmt.Println("Сервер запускается")
	err := http.ListenAndServe(":7540", nil)
	if err != nil {
		fmt.Printf("Ошибка сервера: %s\n", err.Error())
		return
	}

}
