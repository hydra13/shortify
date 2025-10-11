package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

const KeyLength = 8

var repository map[string]string

func generateShortID(url string) string {
	hash := sha256.Sum256([]byte(url))
	encoded := base64.URLEncoding.EncodeToString(hash[:])

	return encoded[:KeyLength]
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		fmt.Printf("Incorrect content type in request: %v\n", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error read request body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Осознанно не делаю проверки на то что получен URL
	// TODO: добавить проверку на длину строки body

	url := string(body)
	key := generateShortID(url)
	repository[key] = url
	shortURL := fmt.Sprintf(`http://localhost:8080/%s`, key)

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	// Длина пути должна быть равна длине ключа + `/`
	if len(r.URL.Path) != KeyLength+1 {
		fmt.Printf("Incorrect request path: %v\n", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key := r.URL.Path[1:]
	url, found := repository[key]

	if !found {
		fmt.Printf("Url not found: %v\n", key)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func wildcardHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		postHandler(w, r)
	case http.MethodGet:
		getHandler(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func main() {
	repository = make(map[string]string)

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, wildcardHandler)

	http.ListenAndServe(":8080", mux)
}
