package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

/*
Tables:
create table users (id serial primary key, username text, password text, email text);
*/

type Config struct {
	DbHost string
	DbPort int
	DbName string
	DbUser string
	DbPass string
}

type App struct {
	Value  string
	Config *Config
	Db     *sql.DB
}

type User struct {
	ID       int
	Username string
	Password string
	Email    string
}

func (app *App) initDb() error {
	sqlCred := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		app.Config.DbHost, app.Config.DbPort, app.Config.DbUser, app.Config.DbPass, app.Config.DbName)
	db, err := sql.Open("postgres", sqlCred)
	if err != nil {
		return err
	}
	app.Db = db
	return nil
}

func (app *App) sayHello(name string) (string, error) {
	if len(name) < 4 {
		return "", errors.New("too-short")
	}
	result := "Hello, " + name + " " + app.Value
	return result, nil
}

func (app *App) getUsers(w http.ResponseWriter, req *http.Request) {
	rows, err := app.Db.Query("SELECT id, username, password, email FROM users")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal-server-error"))
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err = rows.Scan(
			&user.ID,
			&user.Username,
			&user.Password,
			&user.Email,
		)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal-server-error"))
			return
		}
		users = append(users, user)
	}

	jsonByte, err := json.Marshal(users)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal-server-error"))
		return
	}
	w.Write([]byte(string(jsonByte)))
}

func (app *App) getUser(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	var user User
	err := app.Db.QueryRow(
		`SELECT id, username, password, email FROM users WHERE id=$1`,
		id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
	)
	if err != nil {
		// err not found
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal-server-error"))
		return
	}

	jsonByte, err := json.Marshal(user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal-server-error"))
		return
	}
	w.Write([]byte(string(jsonByte)))
}

func main() {
	config := &Config{
		DbHost: "localhost",
		DbPort: 5432,
		DbName: "kelas-go",
		DbUser: "postgres",
		DbPass: "db",
	}

	app := &App{
		Value:  "value",
		Config: config,
	}

	err := app.initDb()
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "hello\n")
	})

	r.HandleFunc("/users", app.getUsers).Methods("GET")
	r.HandleFunc("/user/{id}", app.getUser).Methods("GET")

	http.Handle("/", r)
	log.Println("Running on port 8000")
	http.ListenAndServe(":8000", nil)
}
