package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/rs/cors"
	"github.com/spf13/pflag"
	"goji.io"
	"goji.io/pat"
	"google.golang.org/api/option"
)

var projectID = "gifted-outrider-244523"
var client *storage.Client

//SayHello ...
func SayHello(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("Service is working for SayHello...Received rpc from client \n")
	defer recoverPanic()
	// name := fmt.Sprintf("golang-example-buckets-%d", time.Now().Unix())
	if err := deleteBucketGCS("golang-example-buckets-1571989011"); err != nil {
		panic(err)
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "Hello  service is working..."})
}

func createBucketGCS(bucketName string) error {
	ctx := context.Background()
	// [START create_bucket]
	if err := client.Bucket(bucketName).Create(ctx, projectID, nil); err != nil {
		panic(err)
	}
	// [END create_bucket]
	return nil
}
func deleteBucketGCS(bucketName string) error {
	ctx := context.Background()
	// [START delete_bucket]
	if err := client.Bucket(bucketName).Delete(ctx); err != nil {
		return err
	}
	// [END delete_bucket]
	return nil
}

func recoverPanic() {
	if err := recover(); err != nil {
		fmt.Println("There's something wrong:", err)
	}
}

func writeToBucketGCS(file []byte, fileName string) error {

	ctx := context.Background()
	bucket := client.Bucket("golang-example-buckets-1571989011")
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
	defer recoverPanic()
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

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		respondWithError(w, http.StatusUnsupportedMediaType, err.Error())
		return
	}
	defer recoverPanic()
	if err := writeToBucketGCS(buf.Bytes(), handler.Filename); err != nil {
		panic(err)
	}
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
	fmt.Println("GCS Service Started")
	ctx := context.Background()
	var err error
	client, err = storage.NewClient(ctx, option.WithCredentialsFile("My First Project-cd720c50273d.json"))
	if err != nil {
		panic(err)
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
