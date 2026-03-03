package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Biography string    `json:"biography"`
}

type application struct {
	mu   sync.RWMutex
	data map[uuid.UUID]User
}

func (app *application) FindAll() []User {
	app.mu.RLock()
	defer app.mu.RUnlock()

	users := make([]User, 0, len(app.data))
	for _, user := range app.data {
		users = append(users, user)
	}
	return users
}

func (app *application) FindById(id uuid.UUID) (*User, bool) {
	app.mu.RLock()
	defer app.mu.RUnlock()

	user, exists := app.data[id]
	if !exists {
		return nil, false
	}
	return &user, true
}

func (app *application) Insert(newUser User) User {
	app.mu.Lock()
	defer app.mu.Unlock()

	newUser.ID = uuid.New()
	app.data[newUser.ID] = newUser
	return newUser
}

func (app *application) Update(id uuid.UUID, updates User) (*User, bool) {
	app.mu.Lock()
	defer app.mu.Unlock()

	user, exists := app.data[id]
	if !exists {
		return nil, false
	}

	user.FirstName = updates.FirstName
	user.LastName = updates.LastName
	user.Biography = updates.Biography

	app.data[id] = user
	return &user, true
}

func (app *application) Delete(id uuid.UUID) (*User, bool) {
	app.mu.Lock()
	defer app.mu.Unlock()

	user, exists := app.data[id]
	if !exists {
		return nil, false
	}

	delete(app.data, id)
	return &user, true
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func validateUser(u User) error {
	if len(u.FirstName) < 2 || len(u.FirstName) > 20 {
		return fmt.Errorf("first_name must be between 2 and 20 characters")
	}
	if len(u.LastName) < 2 || len(u.LastName) > 20 {
		return fmt.Errorf("last_name must be between 2 and 20 characters")
	}
	if len(u.Biography) < 20 || len(u.Biography) > 450 {
		return fmt.Errorf("biography must be between 20 and 450 characters")
	}
	return nil
}

func (app *application) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validateUser(input); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	createdUser := app.Insert(input)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
}

func (app *application) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users := app.FindAll()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func (app *application) handleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	user, exists := app.FindById(id)
	if !exists {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (app *application) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var input User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validateUser(input); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	updatedUser, exists := app.Update(id, input)
	if !exists {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}

func (app *application) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	deletedUser, exists := app.Delete(id)
	if !exists {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(deletedUser)
}

func main() {
	app := &application{
		data: make(map[uuid.UUID]User),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/users", app.handleCreateUser)
	mux.HandleFunc("GET /api/users", app.handleListUsers)
	mux.HandleFunc("GET /api/users/{id}", app.handleGetUser)
	mux.HandleFunc("PUT /api/users/{id}", app.handleUpdateUser)
	mux.HandleFunc("DELETE /api/users/{id}", app.handleDeleteUser)

	fmt.Println("Server running on port 8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
