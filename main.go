package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Todo struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	Completed   bool   `json:"completed"`
}

func main() {
	// Load env
	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatal("Error loading .env file")
	}

	// Access env
	port := os.Getenv("PORT")

	////////////

	// // Open db connection
	// dbUser := os.Getenv("DB_USER")
	// dbPass := os.Getenv("DB_PASS")
	// dbHost := os.Getenv("DB_HOST")
	// dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/simple_todo", dbUser, dbPass, dbHost)
	// db, err := sql.Open("mysql", dsn)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close() // close db when main function ends

	// pingErr := db.Ping()
	// if pingErr != nil {
	// 	fmt.Println(pingErr)
	// }

	////////////

	// Initialize router
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	// Define CORS
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
	}))

	// Define API endpoints
	router.Get("/api/todos", getTodos)
	router.Post("/api/todos", createTodo)
	router.Put("/api/todos/{id}", updateTodo)
	router.Delete("/api/todos/{id}", deleteTodo)

	// Start server
	log.Printf("Server starting at port %s", port)
	portString := fmt.Sprintf(":%s", port)
	log.Fatal(http.ListenAndServe(portString, router))
}

func setupDB() *sql.DB {
	// Access env
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/simple_todo", dbUser, dbPass, dbHost)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	// Query all todos from db
	db := setupDB()
	rows, err := db.Query("SELECT id, description, created_at, completed FROM todos")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Iterate over result and build todo slice
	todos := []Todo{}
	for rows.Next() {
		todo := Todo{}
		err := rows.Scan(&todo.ID, &todo.Description, &todo.CreatedAt, &todo.Completed)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	// Convert todo slice to JSON and send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	// Parse request body into a Todo struct
	var todo Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Insert new todo into db
	db := setupDB()
	_, err = db.Exec("INSERT INTO todos (id, description, created_at, completed) VALUES (?, ?, NOW(), ?)", todo.ID, todo.Description, todo.Completed)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	// Extract todo ID from URL
	id := r.URL.Path[len("/api/todos/"):]
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Delete the todo from database
	db := setupDB()
	_, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	// Extract todo ID from URL
	id := r.URL.Path[len("/api/todos/"):] // need better way to get id from url
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse request body into todo
	var updatedTodo Todo
	err := json.NewDecoder(r.Body).Decode(&updatedTodo)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Update task in db
	db := setupDB()
	_, err = db.Exec("UPDATE todos SET description = ?, completed = ? WHERE id = ?", updatedTodo.Description, updatedTodo.Completed, id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
