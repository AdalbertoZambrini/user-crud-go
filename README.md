# Go Users API 🐹 ![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=flat&logo=go&logoColor=white)

A simple and lightweight RESTful API for managing users, built entirely with Go. 

This project was created to practice Go fundamentals, HTTP routing, and concurrent in-memory data management without relying on heavy web frameworks.

## 🚀 Features

* **CRUD Operations:** Create, Read, Update, and Delete users.
* **In-Memory Storage:** Uses a Go map protected by a `sync.RWMutex` for safe, concurrent access.
* **Standard Library:** Built utilizing Go's native `net/http` package (taking advantage of Go 1.22+ routing features).
* **UUIDs:** Integrates `google/uuid` for generating unique user IDs.

## 📋 Prerequisites

* [Go](https://go.dev/dl/) version **1.22** or higher (required for the new `http.ServeMux` path wildcards).

## 🛠️ Getting Started

1. **Clone the repository:**
   ```bash
   git clone [https://github.com/yourusername/go-users-api.git](https://github.com/yourusername/go-users-api.git)
   cd go-users-api


2. **Install the dependencies:**
```bash
go mod tidy

```


3. **Run the server:**
```bash
go run main.go

```


*The server will start running on `http://localhost:8080`.*

## 🛣️ API Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| `POST` | `/api/users` | Create a new user |
| `GET` | `/api/users` | Retrieve a list of all users |
| `GET` | `/api/users/{id}` | Retrieve a specific user by ID |
| `PUT` | `/api/users/{id}` | Update a specific user by ID |
| `DELETE` | `/api/users/{id}` | Delete a specific user by ID |

### 📄 User JSON Object Example

When creating or updating a user, send a JSON body like this:

```json
{
  "first_name": "John",
  "last_name": "Doe",
  "biography": "Software developer currently learning Go."
}

```
