package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/an2kab/sprint-final/internal/api"
	"github.com/an2kab/sprint-final/internal/storage"
	"github.com/go-chi/chi/v5"
)

func NewRouter() (*api.RoutDb, error) {
	// Подключение к БД
	db, err := storage.NewStorage()
	if err != nil {
		fmt.Printf("Ошибка подключения к БД: %s\n", err.Error())
		return nil, err
	}
	return &api.RoutDb{
		DB: db,
	}, nil
}

// Создается сервер с портом подключения 7540, а также подключение к БД и обработчикам через роутер
func NewServer(r *api.RoutDb) *chi.Mux {

	mx := chi.NewRouter()

	// Подключения к обработчикам
	mx.Get("/api/nextdate", r.NextDateHandler)
	mx.Post("/api/task", r.TaskHandler)
	mx.Get("/api/task", r.TaskHandler)
	mx.Put("/api/task", r.TaskHandler)
	mx.Post("/api/task/done", r.TaskHandler)
	mx.Get("/api/tasks", r.TasksHandler)

	// Запуск сервера
	dir, _ := os.Executable()
	mx.Handle("/*", http.FileServer(http.Dir(filepath.Join(filepath.Dir(dir), "web"))))
	fmt.Println("Сервер запускается")
	err := http.ListenAndServe(":7540", mx)
	if err != nil {
		fmt.Printf("Ошибка сервера: %s\n", err.Error())
		return nil
	}
	return mx
}
