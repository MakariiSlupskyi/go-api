package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var (
	DATABASE_URL, DB_DRIVER, PORT string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("could not load env")
	}
	DATABASE_URL = os.Getenv("DATABASE_URL")
	DB_DRIVER = os.Getenv("DB_DRIVER")
	PORT = os.Getenv("PORT")
}

func DBClient() (*sql.DB, error) {
	db, err := sql.Open(DB_DRIVER, DATABASE_URL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	fmt.Println("connected to database")
	return db, nil
}

type Server struct {
	Router *chi.Mux
	DB     *sql.DB
}

func CreateServer(db *sql.DB) *Server {
	server := &Server{
		Router: chi.NewRouter(),
		DB:     db,
	}
	return server
}

func Greet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Glad to see you!"))
}

func (s *Server) MountHandlers() {
	s.Router.Get("/greet", Greet)

	todosRouter := chi.NewRouter()
	todosRouter.Group(func(r chi.Router) {
		r.Get("/", s.GetTodos)
		r.Post("/", s.AddTodo)
	})

	s.Router.Mount("/todos", todosRouter)
}

type Todo struct {
	Id        int       `json:"id"`
	Task      string    `json:"task"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TodoRequestBody struct {
	Task      string `json:"task"`
	Completed bool   `json:"completed"`
}

var Todos []*Todo

func (s *Server) AddTodo(w http.ResponseWriter, r *http.Request) {
	todo := new(TodoRequestBody)
	if err := json.NewDecoder(r.Body).Decode(todo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("incorrect todo entered"))
		return
	}

	query := `INSERT INTO Todos (task, completed) VALUES (?, ?)`
	_, err := s.DB.Exec(query, todo.Task, todo.Completed)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error accured"))
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Todo added!!"))
}

func (s *Server) GetTodos(w http.ResponseWriter, r *http.Request) {
	query := `SELECT * FROM Todos ORDER BY created_at DESC`

	rows, err := s.DB.Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error accured"))
	}

	res := []Todo{}

	for rows.Next() {
		var id int
		var task string
		var completed bool
		var createdAt, updatedAt time.Time

		err = rows.Scan(&id, &task, &completed, &createdAt, &updatedAt)
		if err != nil {
			log.Fatalln("could not parse todo:", err)
		}

		res = append(res, Todo{
			Id:        id,
			Task:      task,
			Completed: completed,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func main() {
	db, err := DBClient()
	if err != nil {
		log.Fatalln("could not connect to database:", err)
	}
	server := CreateServer(db)
	server.MountHandlers()

	fmt.Println("server running on port:5000")
	http.ListenAndServe(":5000", server.Router)
}
