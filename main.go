package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Biography string    `json:"biography"`
}

type application struct {
	dbPool *sqlitex.Pool
}

func (app *application) FindAll() []User {
	users, err := app.findAll(context.Background())
	if err != nil {
		log.Printf("find all users: %v", err)
		return []User{}
	}
	return users
}

func (app *application) FindById(id uuid.UUID) (*User, bool) {
	user, exists, err := app.findByID(context.Background(), id)
	if err != nil {
		log.Printf("find user by id %s: %v", id, err)
		return nil, false
	}
	return user, exists
}

func (app *application) Insert(newUser User) User {
	createdUser, err := app.insert(context.Background(), newUser)
	if err != nil {
		log.Printf("insert user: %v", err)
		return User{}
	}
	return createdUser
}

func (app *application) Update(id uuid.UUID, updates User) (*User, bool) {
	user, exists, err := app.update(context.Background(), id, updates)
	if err != nil {
		log.Printf("update user %s: %v", id, err)
		return nil, false
	}
	return user, exists
}

func (app *application) Delete(id uuid.UUID) (*User, bool) {
	user, exists, err := app.delete(context.Background(), id)
	if err != nil {
		log.Printf("delete user %s: %v", id, err)
		return nil, false
	}
	return user, exists
}

func openDatabase(ctx context.Context) (*sqlitex.Pool, error) {
	pool, err := sqlitex.NewPool("file:users.db", sqlitex.PoolOptions{
		PoolSize: 10,
		PrepareConn: func(conn *sqlite.Conn) error {
			conn.SetBusyTimeout(5 * time.Second)
			return nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite pool: %w", err)
	}

	conn, err := pool.Take(ctx)
	if err != nil {
		_ = pool.Close()
		return nil, fmt.Errorf("take sqlite connection: %w", err)
	}
	defer pool.Put(conn)

	const createUsersTableSQL = `
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    biography TEXT NOT NULL
);
`
	if err := sqlitex.ExecuteScript(conn, createUsersTableSQL, nil); err != nil {
		_ = pool.Close()
		return nil, fmt.Errorf("create users table: %w", err)
	}

	return pool, nil
}

func (app *application) findAll(ctx context.Context) ([]User, error) {
	conn, err := app.dbPool.Take(ctx)
	if err != nil {
		return nil, fmt.Errorf("take sqlite connection: %w", err)
	}
	defer app.dbPool.Put(conn)

	users := make([]User, 0)
	err = sqlitex.Execute(conn, "SELECT id, first_name, last_name, biography FROM users ORDER BY rowid", &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			id, err := uuid.Parse(stmt.ColumnText(0))
			if err != nil {
				return fmt.Errorf("parse user id from database: %w", err)
			}

			users = append(users, User{
				ID:        id,
				FirstName: stmt.ColumnText(1),
				LastName:  stmt.ColumnText(2),
				Biography: stmt.ColumnText(3),
			})
			return nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}

	return users, nil
}

func fetchUserByID(conn *sqlite.Conn, id uuid.UUID) (*User, bool, error) {
	var user *User

	err := sqlitex.Execute(conn, "SELECT id, first_name, last_name, biography FROM users WHERE id = ?1", &sqlitex.ExecOptions{
		Args: []any{id.String()},
		ResultFunc: func(stmt *sqlite.Stmt) error {
			parsedID, err := uuid.Parse(stmt.ColumnText(0))
			if err != nil {
				return fmt.Errorf("parse user id from database: %w", err)
			}

			user = &User{
				ID:        parsedID,
				FirstName: stmt.ColumnText(1),
				LastName:  stmt.ColumnText(2),
				Biography: stmt.ColumnText(3),
			}
			return nil
		},
	})
	if err != nil {
		return nil, false, fmt.Errorf("query user by id: %w", err)
	}

	if user == nil {
		return nil, false, nil
	}
	return user, true, nil
}

func (app *application) findByID(ctx context.Context, id uuid.UUID) (*User, bool, error) {
	conn, err := app.dbPool.Take(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("take sqlite connection: %w", err)
	}
	defer app.dbPool.Put(conn)

	return fetchUserByID(conn, id)
}

func (app *application) insert(ctx context.Context, newUser User) (User, error) {
	conn, err := app.dbPool.Take(ctx)
	if err != nil {
		return User{}, fmt.Errorf("take sqlite connection: %w", err)
	}
	defer app.dbPool.Put(conn)

	newUser.ID = uuid.New()
	err = sqlitex.Execute(conn, "INSERT INTO users (id, first_name, last_name, biography) VALUES (?1, ?2, ?3, ?4)", &sqlitex.ExecOptions{
		Args: []any{
			newUser.ID.String(),
			newUser.FirstName,
			newUser.LastName,
			newUser.Biography,
		},
	})
	if err != nil {
		return User{}, fmt.Errorf("insert user: %w", err)
	}

	return newUser, nil
}

func (app *application) update(ctx context.Context, id uuid.UUID, updates User) (*User, bool, error) {
	conn, err := app.dbPool.Take(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("take sqlite connection: %w", err)
	}
	defer app.dbPool.Put(conn)

	err = sqlitex.Execute(conn, "UPDATE users SET first_name = ?1, last_name = ?2, biography = ?3 WHERE id = ?4", &sqlitex.ExecOptions{
		Args: []any{
			updates.FirstName,
			updates.LastName,
			updates.Biography,
			id.String(),
		},
	})
	if err != nil {
		return nil, false, fmt.Errorf("update user: %w", err)
	}

	if conn.Changes() == 0 {
		return nil, false, nil
	}

	updatedUser := &User{
		ID:        id,
		FirstName: updates.FirstName,
		LastName:  updates.LastName,
		Biography: updates.Biography,
	}
	return updatedUser, true, nil
}

func (app *application) delete(ctx context.Context, id uuid.UUID) (*User, bool, error) {
	conn, err := app.dbPool.Take(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("take sqlite connection: %w", err)
	}
	defer app.dbPool.Put(conn)

	user, exists, err := fetchUserByID(conn, id)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}

	err = sqlitex.Execute(conn, "DELETE FROM users WHERE id = ?1", &sqlitex.ExecOptions{
		Args: []any{id.String()},
	})
	if err != nil {
		return nil, false, fmt.Errorf("delete user: %w", err)
	}

	if conn.Changes() == 0 {
		return nil, false, nil
	}
	return user, true, nil
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

	createdUser, err := app.insert(r.Context(), input)
	if err != nil {
		log.Printf("create user: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
}

func (app *application) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.findAll(r.Context())
	if err != nil {
		log.Printf("list users: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

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

	user, exists, err := app.findByID(r.Context(), id)
	if err != nil {
		log.Printf("get user %s: %v", id, err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

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

	updatedUser, exists, err := app.update(r.Context(), id, input)
	if err != nil {
		log.Printf("update user %s: %v", id, err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

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

	deletedUser, exists, err := app.delete(r.Context(), id)
	if err != nil {
		log.Printf("delete user %s: %v", id, err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if !exists {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(deletedUser)
}

func main() {
	dbPool, err := openDatabase(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := dbPool.Close(); err != nil {
			log.Printf("close sqlite pool: %v", err)
		}
	}()

	app := &application{dbPool: dbPool}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/users", app.handleCreateUser)
	mux.HandleFunc("GET /api/users", app.handleListUsers)
	mux.HandleFunc("GET /api/users/{id}", app.handleGetUser)
	mux.HandleFunc("PUT /api/users/{id}", app.handleUpdateUser)
	mux.HandleFunc("DELETE /api/users/{id}", app.handleDeleteUser)

	fmt.Println("Server running on port 8080...")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
