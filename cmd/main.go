package main

import (
	"github.com/an2kab/sprint-final/internal/server"
)

func main() {

	//storage.NewStorage()
	s, err := server.NewRouter()
	if err != nil {
		return
	}
	server.NewServer(s)

}
