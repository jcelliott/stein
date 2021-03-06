package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/beatgammit/stein"
	"github.com/codegangsta/martini"
	log "github.com/jcelliott/lumber"
	"net/http"
	"os"
	"time"
)

var dbType string

func init() {
	flag.StringVar(&dbType, "dbtype", "fs", "database type to use: fs, couchdb")
	flag.Parse()
}

func main() {
	var db DB
	var err error
	switch dbType {
	case "fs":
		db, err = NewFileStore("file_store")
	case "couchdb":
		db, err = NewCouchDB("localhost:5984", "test", "", "")
	default:
		err = fmt.Errorf("Unsupported database type: %s", dbType)
	}

	if err != nil {
		log.Error("Error initializing database: %s", err)
		os.Exit(1)
		return
	}

	m := martini.Classic()
	m.Use(martini.Static("build/web"))
	m.Get("/projects", func() (string, int) {
		projs, err := db.GetProjects()
		if err != nil {
			return err.Error(), 500
		}
		b, _ := json.Marshal(projs)
		return string(b), 200
	})

	m.Get("/projects/:project/tests", func(params martini.Params) (string, int) {
		tests, err := db.GetTests(params["project"])
		if err != nil {
			return err.Error(), 500
		}
		b, _ := json.Marshal(tests)
		return string(b), 200
	})

	m.Post("/projects/:project/tests", func(params martini.Params, r *http.Request) (string, int) {
		id := time.Now().Format(time.RFC3339)
		s, err := stein.Parse(r.Body)
		if err != nil {
			return err.Error(), 500
		}

		err = db.Save(params["project"], id, s)
		if err != nil {
			return err.Error(), 500
		}
		return id, 200
	})
	m.Get("/projects/:project/tests/:test", func(params martini.Params) (string, int) {
		s, err := db.GetTest(params["project"], params["test"])
		if err != nil {
			return err.Error(), 500
		}
		b, _ := json.Marshal(s)
		return string(b), 200
	})
	m.Run()
}
