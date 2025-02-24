package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title,omitempty"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type RoutDb struct {
	DB *sql.DB
}

type ResponseTask struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

// JsonError - функция отправки json с текстом ошибки
func JsonError(w http.ResponseWriter, text string) {
	var responseTask ResponseTask
	responseTask.Error = text
	//fmt.Printf("%s: %s\n", text)

	res, err := json.Marshal(responseTask)
	if err != nil {
		http.Error(w, "ошибка сериализации json с ошибкой", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write(res)
}

// NextDateHandler - обработчик повторения задачи раз в год или через определенное количество дней
func (rb *RoutDb) NextDateHandler(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	parseNow, err := time.Parse("20060102", now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := NextDate(parseNow, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(response))
}

// TaskHandler - обработчик добавления задачи, получения информации о задаче, редактирования задачи и удаления задачи
func (rb *RoutDb) TaskHandler(w http.ResponseWriter, r *http.Request) {

	var responseTask ResponseTask
	var task Task
	var parseDate time.Time
	var nextDate string

	switch r.Method {
	// Добавление задачи
	case "POST":
		w.Header().Set("Content-Type", "applicatiom/json; charset=UTF-8")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		//defer r.Body.Close()

		//Десериализация тела запроса в структуру
		err = json.Unmarshal(body, &task)
		if err != nil {
			http.Error(w, "ошибка десериализации JSON", http.StatusBadRequest)
		}
		fmt.Printf("Task: %v\n", task)

		if task.Title == "" {
			JsonError(w, "не указан заголовок задачи")
			//http.Error(w, "не указан заголовок задачи", http.StatusBadRequest)
			return
		}

		if task.Date == "" {
			task.Date = time.Now().Format("20060102")
		} else {
			parseDate, err = time.Parse("20060102", task.Date)
			if err != nil {
				JsonError(w, "дата представлена в формате, отличном от 20060102")
				//http.Error(w, "дата представлена в формате, отличном от 20060102", http.StatusBadRequest)
				return
			}

			if task.Date == time.Now().Format("20060102") {
				task.Date = time.Now().Format("20060102")
			} else if parseDate.Before(time.Now()) {
				switch {
				case task.Repeat == "":
					task.Date = time.Now().Format("20060102")
				default:
					nextDate, err = NextDate(time.Now(), task.Date, task.Repeat)
					if err != nil {
						JsonError(w, "правило повторения указано в неправильном формате")
						//http.Error(w, "правило повторения указано в неправильном формате", http.StatusBadRequest)
						return
					}
					task.Date = nextDate
				}
			}

		}
		// fmt.Println("Запись в БД...")
		// fmt.Printf("date: %s\n", task.Date)
		// fmt.Printf("title: %s\n", task.Title)
		// fmt.Printf("comment: %s\n", task.Comment)
		// fmt.Printf("repeat: %s\n", task.Repeat)
		resDb, err := rb.DB.Exec(`INSERT INTO scheduler	(date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)`, sql.Named("date", task.Date), sql.Named("title", task.Title), sql.Named("comment", task.Comment), sql.Named("repeat", task.Repeat))
		if err != nil {
			http.Error(w, "ошибка добавления в БД", http.StatusInternalServerError)
			return
		}
		fmt.Println("Запись в БД выпонена")

		id, err := resDb.LastInsertId()
		if err != nil {
			http.Error(w, "ошибка получения id", http.StatusInternalServerError)
			return
		}
		responseTask.ID = strconv.Itoa(int(id))
		fmt.Printf("id: %s\n", responseTask.ID)

		res, err := json.MarshalIndent(responseTask, "", " ")
		if err != nil {
			http.Error(w, "ошибка сериализации", http.StatusInternalServerError)
			return
		}
		//fmt.Println(string(resp))
		w.Write(res)

	// Получение информации о задаче
	case "GET":
		q := r.URL.Query()
		idString := q.Get("id")
		idInteger, err := strconv.Atoi(idString)
		if err != nil {
			http.Error(w, "Задача не найдена", http.StatusBadRequest)
			return
		}

		task := Task{}

		//обращение к БД
		row := rb.DB.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :idInt`, sql.Named("idInt", idInteger))
		err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			http.Error(w, "ошибка БД", http.StatusInternalServerError)
			return
		}
		res, err := json.MarshalIndent(task, "", " ")
		if err != nil {
			fmt.Printf("ошибка сериализации ответа: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(res)

	// Редактирование задачи
	case "PUT":
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		task := Task{}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "ошибка тела запроса", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		//десериализация тела запроса в структуру
		err = json.Unmarshal(body, &task)
		if err != nil {
			JsonError(w, "ошибка десериализации JSON")
			//http.Error(w, "ошибка десериализации JSON", http.StatusBadRequest)
			return
		}

		idInteger, err := strconv.Atoi(task.ID)
		if err != nil {
			JsonError(w, "ошибка в конвертированиии id")
			//http.Error(w, "ошибка в конвертированиии id", http.StatusInternalServerError)
			return
		}

		if task.Title == "" {
			JsonError(w, "не указан заголовок задачи")
			//http.Error(w, "не указан заголовок задачи", http.StatusBadRequest)
			return
		}

		if task.Date == "" {
			task.Date = time.Now().Format("20060102")
		} else {
			parseDate, err = time.Parse("20060102", task.Date)
			if err != nil {
				JsonError(w, "ошибка формата времени")
				//http.Error(w, "ошибка формата времени", http.StatusBadRequest)
				return
			}

			if task.Date == time.Now().Format("20060102") {
				task.Date = time.Now().Format("20060102")
			} else if parseDate.Before(time.Now()) {
				switch {
				case task.Repeat == "":
					task.Date = time.Now().Format("20060102")
				default:
					nextDate, err = NextDate(time.Now(), task.Date, task.Repeat)
					if err != nil {
						JsonError(w, "ошибка вычисления следующей даты")
						//http.Error(w, "ошибка вычисления следующей даты", http.StatusInternalServerError)
						return
					}
					task.Date = nextDate
				}
			}

		}
		fmt.Println("Запись в БД...")

		// проверка наличия id
		row := rb.DB.QueryRow(`SELECT id FROM scheduler WHERE id = :id`, sql.Named("id", idInteger))

		var i int
		err = row.Scan(&i)
		if err != nil {
			JsonError(w, "id отсутствует в базе")
			//http.Error(w, "id отсутствует в базе", http.StatusInternalServerError)
			return
		}

		// обновление задачи
		_, err = rb.DB.Exec(`UPDATE scheduler SET date = :date,	title = :title,	comment = :comment,	repeat = :repeat WHERE id = :id`, sql.Named("date", task.Date), sql.Named("title", task.Title), sql.Named("comment", task.Comment), sql.Named("repeat", task.Repeat), sql.Named("id", idInteger))
		if err != nil {
			JsonError(w, "ошибка обновления задачи")
			//http.Error(w, "ошибка обновления задачи", http.StatusInternalServerError)
			return
		}
		fmt.Println("Запись в БД выполнена успешно")

		var empty Task

		res, err := json.Marshal(empty)
		if err != nil {
			//fmt.Printf("ошибка сериализации ответа: %s\n", err.Error())
			JsonError(w, "ошибка сериализации ответа")
			//http.Error(w, "ошибка сериализации ответа:", http.StatusInternalServerError)
			return
		}
		w.Write(res)

	//Удаление задачи
	case "DELETE":
		q := r.URL.Query()
		idString := q.Get("id")
		idInteger, err := strconv.Atoi(idString)
		if err != nil {
			JsonError(w, "Задача не найдена")
			//http.Error(w, "Задача не найдена", http.StatusBadRequest)
			return
		}

		//удаление задачи по id
		_, err = rb.DB.Exec(`DELETE FROM scheduler WHERE id = :idInt`, sql.Named("idInt", idInteger))
		if err != nil {
			JsonError(w, "ошибка БД")
			//http.Error(w, "ошибка БД", http.StatusInternalServerError)
			return
		}

		var empty ResponseTask
		res, err := json.Marshal(empty)
		if err != nil {
			//JsonError(w, "Ошибка сериализации")
			http.Error(w, "Ошибка сериализации", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(res)
	}
}

// TasksHandler - обработчик получения списка ближайших задач
func (rb *RoutDb) TasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var responseTask ResponseTask
	tasks := make([]Task, 0)
	now := time.Now().Format("20060102")
	//обращение к БД
	row, err := rb.DB.Query(`SELECT id, date, title, comment, repeat FROM scheduler WHERE date >= :now	ORDER BY date`, sql.Named("now", now))
	if err != nil {
		responseTask.Error = err.Error()

		res, err := json.Marshal(responseTask)
		if err != nil {
			http.Error(w, "ошибка сериализации", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res)
		return

	}
	defer row.Close()

	// Парсинг строк из БД
	for row.Next() {
		temporary := Task{}

		err := row.Scan(&temporary.ID, &temporary.Date, &temporary.Title, &temporary.Comment, &temporary.Repeat)
		if err != nil {
			responseTask.Error = err.Error()

			res, err := json.Marshal(responseTask)
			if err != nil {
				http.Error(w, "ошибка сериализации", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(res)
			return
		}
		tasks = append(tasks, temporary)
	}

	if err := row.Err(); err != nil {
		responseTask.Error = err.Error()

		res, err := json.Marshal(responseTask)
		if err != nil {
			http.Error(w, "ошибка сериализации", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res)
		return
	}
	m := make(map[string][]Task)
	m["tasks"] = tasks
	res, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		http.Error(w, "ошибка сериализации", http.StatusInternalServerError)
		return
	}

	w.Write(res)
}

// DoneTaskHandler - обработчик выполнения (завершения) задачи
func (rb *RoutDb) DoneTaskHandler(w http.ResponseWriter, r *http.Request) {

	var empty ResponseTask

	q := r.URL.Query()
	idString := q.Get("id")
	idInteger, err := strconv.Atoi(idString)
	if err != nil {
		JsonError(w, "Задача не найдена")
		//http.Error(w, "Задача не найдена", http.StatusBadRequest)
		return
	}

	task := Task{}

	//обращение к БД
	row := rb.DB.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :idInteger`, sql.Named("idInteger", idInteger))
	err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		JsonError(w, "ошибка БД")
		//http.Error(w, "ошибка БД", http.StatusInternalServerError)
		return
	}
	// Проверка на повторяемость задачи
	if task.Repeat == "" {
		_, err := rb.DB.Exec(`DELETE FROM scheduler	WHERE id = :idInteger`, sql.Named("idInteger", idInteger))
		if err != nil {
			JsonError(w, "ошибка БД")
			//http.Error(w, "ошибка БД", http.StatusInternalServerError)
			return
		}

		res, err := json.Marshal(empty)
		if err != nil {
			JsonError(w, "ошибка сериализации ответа")
			//http.Error(w, "ошибка сериализации ответа", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(res)

	} else {
		//если задача повторяется, необходимо изменить дату следующего выполнения задачи
		now := time.Now()

		newDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			JsonError(w, "непправильный формат")
			//http.Error(w, "непправильный формат", http.StatusBadRequest)
			return
		}
		task.Date = newDate

		//вносим изменения в БД
		_, err = rb.DB.Exec(`UPDATE scheduler SET date = :date WHERE id = :id`, sql.Named("date", task.Date), sql.Named("id", task.ID))
		if err != nil {
			JsonError(w, "ошибка БД")
			//http.Error(w, "ошибка БД", http.StatusInternalServerError)
			return
		}

		res, err := json.Marshal(empty)
		if err != nil {
			http.Error(w, "ошибка сериализации ответа", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(res)

	}

}
