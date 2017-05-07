package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Todo struct {
	ID    int
	Title string
	Body  string
}

type Todos struct {
	mux   sync.Mutex
	todos []Todo
	seq   int
}

func (t *Todos) Count() int {
	t.mux.Lock()
	defer t.mux.Unlock()

	return len(t.todos)
}

func NewTodos() *Todos {
	ret := Todos{
		mux:   sync.Mutex{},
		todos: []Todo{},
	}

	return &ret
}

func (t *Todos) Add(todo *Todo) {
	t.mux.Lock()
	defer t.mux.Unlock()

	todo.ID, t.seq = t.seq, t.seq+1
	t.todos = append(t.todos, *todo)
}

func (t *Todos) All() []Todo {
	return t.todos
}

func (t *Todos) Remove(id int) bool {
	t.mux.Lock()
	defer t.mux.Unlock()

	for idx, v := range t.todos {
		if v.ID == id {
			t.todos = append(t.todos[:idx], t.todos[idx+1:]...)
			return true
		}
	}

	return false
}

var todos *Todos

func main() {

	todos = NewTodos()
	ret.Add(&Todo{Title: "Title 1", Body: "Body 1"})
	ret.Add(&Todo{Title: "Title 2", Body: "Body 2"})
	ret.Add(&Todo{Title: "Title 3", Body: "Body 3"})

	r := mux.NewRouter()

	r.HandleFunc("/", renderPage).Methods("GET")
	r.HandleFunc("/todos", getTodos).Methods("GET")
	r.HandleFunc("/todos", createTodo).Methods("POST")
	r.HandleFunc("/todos/{id:[0-9]+}", removeTodo).Methods("DELETE")

	http.Handle("/", r)
	http.ListenAndServe(":3000", nil)
}

func renderPage(w http.ResponseWriter, r *http.Request) {

	b, err := ioutil.ReadFile("static/index.html")
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, []byte(fmt.Sprintf("%v", err)))
	} else {
		writeHTMLResponse(w, http.StatusOK, b)
	}
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, []byte(fmt.Sprintf("%v", err)))
		return
	}

	todo := &Todo{}
	err = json.Unmarshal(b, todo)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, []byte("Invalid body"))
	} else {
		todos.Add(todo)
		b, err = json.Marshal(todo)
		if err != nil {
			todos.Remove(todo.ID)
			writeErrorResponse(w, http.StatusInternalServerError, []byte("Internal server error"))
		} else {
			writeJSONResponse(w, http.StatusOK, b)
		}
	}
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(todos.All())
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, []byte("Internal server error"))
	} else {
		writeJSONResponse(w, http.StatusOK, b)
	}
}

func removeTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err == nil {
		removed := todos.Remove(id)
		if removed {
			writeJSONResponse(w, http.StatusNoContent, []byte{})
		} else {
			writeErrorResponse(w, http.StatusNotFound, []byte("Could not find any todo with the id "+vars["id"]))
		}
	} else {
		writeErrorResponse(w, http.StatusBadRequest, []byte("Invalid id "+vars["id"]))
	}
}

func writeJSONResponse(w http.ResponseWriter, code int, response []byte) {
	writeResponse(w, code, "application/json", response)
}

func writeHTMLResponse(w http.ResponseWriter, code int, response []byte) {
	writeResponse(w, code, "text/html", response)
}

func writeErrorResponse(w http.ResponseWriter, code int, response []byte) {
	writeResponse(w, code, "text/plain", response)
}

func writeResponse(w http.ResponseWriter, code int, contentType string, response []byte) {
	w.Header().Set("content-type", contentType)
	w.WriteHeader(code)
	w.Write(response)
}
