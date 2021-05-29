package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func getStatus(w http.ResponseWriter, req *http.Request) {
	jsonResponse(w, Message{false, "ok"})
}

func postReading(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errorResponse(w, err)
		return
	}
	var reading Reading
	err = json.Unmarshal(body, &reading)
	if err != nil {
		errorResponse(w, err)
		return
	}
	reading.Timestamp = time.Now().UTC()
	err = logNewReading(&reading)
	if err != nil {
		errorResponse(w, err)
		return
	}
	jsonResponse(w, reading)
}

func getReadings(w http.ResponseWriter, req *http.Request) {
	limit, offset, err := getLimitOffset(req)
	if err != nil {
		errorResponse(w, err)
		return
	}
	readings, err := getOrphanReadingsInRange(limit, offset)
	if err != nil {
		errorResponse(w, err)
		return
	}
	jsonResponse(w, readings)
}

func saveCook(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errorResponse(w, err)
		return
	}
	var cook Cook
	err = json.Unmarshal(body, &cook)
	if err != nil {
		errorResponse(w, err)
		return
	}
	if req.Method == "POST" {
		cook.Created = time.Now().UTC()
		err = createNewCook(&cook)
		if err != nil {
			errorResponse(w, err)
			return
		}
	} else if req.Method == "PUT" {
		vars := mux.Vars(req)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			errorResponse(w, err)
			return
		}
		if id != cook.Id {
			errorResponse(w, errors.New("id mismatch"))
			return
		}
		err = updateCook(&cook)
		if err != nil {
			errorResponse(w, err)
			return
		}
	} else {
		errorResponse(w, errors.New("bad method"))
		return
	}
	jsonResponse(w, cook)
}

func getCooks(w http.ResponseWriter, req *http.Request) {
	limit, offset, err := getLimitOffset(req)
	if err != nil {
		errorResponse(w, err)
		return
	}
	cooks, err := getCooksInRange(limit, offset)
	if err != nil {
		errorResponse(w, err)
		return
	}
	jsonResponse(w, cooks)
}

func getCook(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		errorResponse(w, err)
		return
	}
	cook, err := getCookById(id)
	if err != nil {
		errorResponse(w, err)
		return
	}
	jsonResponse(w, cook)
}

func updateCookReadings(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		errorResponse(w, err)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errorResponse(w, err)
		return
	}
	var update CookReadingUpdate
	err = json.Unmarshal(body, &update)
	if err != nil {
		errorResponse(w, err)
		return
	}
	if len(update.Add) > 0 {
		err = addReadingsToCook(update.Add, id)
		if err != nil {
			errorResponse(w, err)
			return
		}
	}
	if len(update.Remove) > 0 {
		err = removeReadingsFromCook(update.Remove)
		if err != nil {
			errorResponse(w, err)
			return
		}
	}
	jsonResponse(w, Message{false, "ok"})
}

func runHttpServer() {
	r := mux.NewRouter()

	r.HandleFunc("/api", getStatus)
	r.HandleFunc("/api/reading", postReading).Methods("POST")
	r.HandleFunc("/api/reading", getReadings).Methods("GET")
	r.HandleFunc("/api/cook", saveCook).Methods("POST")
	r.HandleFunc("/api/cook", getCooks).Methods("GET")
	r.HandleFunc("/api/cook/{id}", getCook).Methods("GET")
	r.HandleFunc("/api/cook/{id}", saveCook).Methods("PUT")
	r.HandleFunc("/api/cook/{id}/readings", updateCookReadings).Methods("PUT")

	var handler http.Handler = r
	handler = logRequestHandler(handler)

	srv := &http.Server{
		Addr:    os.Getenv("HTTP_SERVER"),
		Handler: handler,
	}
	srv.ListenAndServe()
}
