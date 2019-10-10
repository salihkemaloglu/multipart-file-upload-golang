package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/cors"
	"github.com/spf13/pflag"
	"goji.io"
	"goji.io/pat"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

//SayHello ...
func SayHello(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("Service is working for SayHello...Received rpc from client \n")
	explicit()
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "Hello  service is working..."})
}

type File struct {
	Name               string `json:"Name"`
	ContentType        string `json:"ContentType"`
	ContentLanguage    string `json:"ContentLanguage"`
	CacheControl       string `json:"CacheControl"`
	EventBasedHold     bool   `json:"EventBasedHold"`
	TemporaryHold      bool   `json:"TemporaryHold"`
	PredefinedACL      string `json:"PredefinedACL"`
	Owner              string `json:"Owner"`
	Size               int32  `json:"Size"`
	ContentEncoding    string `json:"ContentEncoding"`
	ContentDisposition string `json:"ContentDisposition"`
	Generation         int16  `json:"Generation"`
	Metageneration     int16  `json:"Metageneration"`
	StorageClass       string `json:"StorageClass"`
	CustomerKeySHA256  string `json:"CustomerKeySHA256"`
	KMSKeyName         string `json:"KMSKeyName"`
	Prefix             string `json:"Prefix"`
	Etag               string `json:"Etag"`
}

// Authenticate to Google Cloud Storage and return handler
func explicit() []File {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("My First Project-cd720c50273d.json"))
	defer recoverPanic()
	if err != nil {
		panic(err)
	}
	it := client.Bucket("gignox_bucker-001").Objects(ctx, nil)
	// var files []File
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		defer recoverPanic()
		if err != nil {
			panic(err)
		}
		fmt.Println(attrs)
		fmt.Println(attrs.Name)
	}
	return nil
}
func recoverPanic() {
	if err := recover(); err != nil {
		fmt.Println("There's something wrong:", err)
	}
}
func writeToCloudStorage(file []byte, fileName string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("My First Project-cd720c50273d.json"))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	bucket := client.Bucket("gignox_bucker-001")

	wc := bucket.Object(fileName).NewWriter(ctx)
	wc.ContentType = "text/plain"

	if _, err := wc.Write(file); err != nil {
		fmt.Println(ctx, "createFile: unable to write data to bucket %q, file %q: %v", bucket, fileName, err)
		return err
	}

	if err := wc.Close(); err != nil {
		fmt.Println(ctx, "createFile: unable to close bucket %q, file %q: %v", bucket, fileName, err)
		return err
	}

	return nil
}

//UploadFile ...
func UploadFile(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("Service is working for UploadFile...Received rpc from client \n")

	file, handler, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusUnsupportedMediaType, err.Error())
		return
	}
	if handler.Size >= 10000000 {
		respondWithError(w, http.StatusNotAcceptable, "File size  can not be bigger than 10 MB!")
		return
	}
	if reqToken := r.Header.Get("Authorization"); reqToken != "JWT" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized user attempt!")
		return
	}

	defer file.Close()
	// copy example

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		respondWithError(w, http.StatusUnsupportedMediaType, err.Error())
		return
	}
	err1 := writeToCloudStorage(buf.Bytes(), handler.Filename)
	_ = err1
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
