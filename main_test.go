package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func BaseTest() *App {
	config := &Config{
		DbHost: "localhost",
		DbPort: 5432,
		DbName: "db",
		DbUser: "postgres",
		DbPass: "password",
	}

	// Migration
	cwd, _ := os.Getwd()
	migrationPath := "file://" + cwd + "/migrations"
	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.DbUser,
		config.DbPass,
		config.DbHost,
		config.DbPort,
		config.DbName,
	)

	// Inisiasi migrasi
	m, err := migrate.New(
		migrationPath,
		url,
	)
	if err != nil {
		log.Fatal(err)
	}
	// Drop semua yang ada
	err = m.Drop()
	if err != nil && err.Error() != "no change" {
		log.Fatal(err)
	}
	// Inisiasi ulang migrasi
	m, err = migrate.New(
		migrationPath,
		url,
	)
	if err != nil {
		log.Fatal(err)
	}
	// Migrasi sampai yang terbaru
	err = m.Up()
	if err != nil && err.Error() != "no change" {
		log.Fatal(err)
	}

	app := &App{
		Value:  "value",
		Config: config,
	}

	err = app.initDb()
	if err != nil {
		log.Fatal(err)
	}

	return app
}

func TestGetUsers(t *testing.T) {
	app := BaseTest()

	request := httptest.NewRequest("GET", "/users", nil)
	response := httptest.NewRecorder()
	handler := http.HandlerFunc(app.getUsers)
	handler.ServeHTTP(response, request)
	bodyBytes, err := ioutil.ReadAll(response.Body)
	assert.Equal(t, nil, err)

	assert.Equal(t, int(200), response.Code)
	users := []User{}
	err = json.Unmarshal(bodyBytes, &users)
	assert.Equal(t, nil, err)
	assert.Equal(t, "piko", users[0].Username)
	assert.Equal(t, "herpiko@gmail.com", users[0].Email)
}

/* TDD, Test Driven Development

CRUD users

- TestCreateUser()
- TestGetUser()
  - create user
  - get user
- TestUpdateUser()
  - create user
  - update user
  - get user
- TestDeleteUser()
  - create user
  - delete user
  - get user
- TestGetUser() // paginasi
  - create user
  - create user
  - create user
  - create user
  - get users (plural) // 4
  - delete user
  - get users (plural) // 3
  - create user // 13 user
  - get users (plural) // page 1, limit 5, expect 5
  - get users (plural) // page 2, limit 5, expect 5
  - get users (plural) // page 3, limit 5, expect 3
*/

func TestCreateUser(t *testing.T) {
	app := BaseTest()

	jsonByte := []byte(`
    {
      "Username":"ananda",
      "Email":"ananda@blengon.in",
      "Password":"YYY"
    }
  `)
	request := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonByte))
	response := httptest.NewRecorder()
	handler := http.HandlerFunc(app.createUser)
	handler.ServeHTTP(response, request)
	bodyBytes, err := ioutil.ReadAll(response.Body)
	log.Println(string(bodyBytes))
	assert.Equal(t, nil, err)
	assert.Equal(t, int(200), response.Code)

	request = httptest.NewRequest("GET", "/users", nil)
	response = httptest.NewRecorder()
	handler = http.HandlerFunc(app.getUsers)
	handler.ServeHTTP(response, request)
	bodyBytes, err = ioutil.ReadAll(response.Body)
	assert.Equal(t, nil, err)
	assert.Equal(t, int(200), response.Code)
	users := []User{}
	err = json.Unmarshal(bodyBytes, &users)
	assert.Equal(t, nil, err)
	assert.Equal(t, int(2), len(users))
	assert.Equal(t, "piko", users[0].Username)
	assert.Equal(t, "herpiko@gmail.com", users[0].Email)
	assert.Equal(t, "ananda", users[1].Username)
	assert.Equal(t, "ananda@blengon.in", users[1].Email)
}

/*
                other-app/module
                   |
                   |
client ---------- app -------------- db
                   |
                   |
                  func

TestXXXX()
   0. Kosongin db
   1. Migrasin db
   2. Inisiasi app
   3. Jalanin UT

TestYYYY()
   2. Migrasin ulang db
   3. Jalanin UT

*/
