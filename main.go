package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

type Server struct {
	Router *chi.Mux
}

func CreateServer() *Server {
	server := &Server{
		Router: chi.NewRouter(),
	}
	return server
}

func Greet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Glad to see you!"))
}

func (server *Server) MountHandlers() {
	server.Router.Get("/greet", Greet)

	todosRouter := chi.NewRouter()
	todosRouter.Group(func(r chi.Router) {
		r.Get("/", GetTodos)
		r.Post("/", AddTodo)
	})

	server.Router.Mount("/todos", todosRouter)
}

type Todo struct {
	Task      string `json:"task"`
	Completed bool   `json:"completed"`
}

var Todos []*Todo

func AddTodo(w http.ResponseWriter, r *http.Request) {
	todo := new(Todo)
	if err := json.NewDecoder(r.Body).Decode(todo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please enter a correct Todo!!"))
		return
	}
	Todos = append(Todos, todo)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Todo added!!"))
}

func GetTodos(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Todos)
}

func main() {
	server := CreateServer()
	server.MountHandlers()

	fmt.Println("server running on port:5000")
	http.ListenAndServe(":5000", server.Router)
}
