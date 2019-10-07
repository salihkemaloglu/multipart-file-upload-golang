package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"io"
	"os"
	"path/filepath"
	"github.com/rs/cors"
	"github.com/spf13/pflag"
	"goji.io"
	"goji.io/pat"
)

//SayHello ...
func SayHello(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("Service is working for SayHello...Received rpc from client \n")
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "Hello  service is working..."})
}


//UploadFile ...
func UploadFile(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("Service is working for UploadFile...Received rpc from client \n")
	
	file, handler, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusUnsupportedMediaType, err.Error())
		return
	}
	if handler.Size >=10000000 {
		respondWithError(w, http.StatusNotAcceptable, "File size  can not be bigger than 10 MB!")
		return
	}
    if 	reqToken := r.Header.Get("Authorization");reqToken!="JWT"{
		respondWithError(w, http.StatusUnauthorized, "Unauthorized user attempt!")
		return
	}

	defer file.Close()
	// copy example
	absPath, _ := filepath.Abs(handler.Filename)
	f, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err.Error())
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("There's something wrong:", err)
			}
		}()
	}
	defer f.Close()
	io.Copy(f, file)
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
func main() {
	pflag.Parse()
	fmt.Println("QC Service is Starting...")
	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/hello"), SayHello)
	mux.HandleFunc(pat.Post("/uploadfile"), UploadFile)
	httpServer := http.Server{
		Addr:    fmt.Sprintf(":%v", 8900),
		Handler: cors.AllowAll().Handler(mux),
	}

	fmt.Printf("server started as http and listen to port: %v \n", 8900)
	if err := httpServer.ListenAndServe(); err != nil {
		fmt.Printf("failed starting http server: %v", err.Error())
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, map[string]string{"error": msg})
}
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}