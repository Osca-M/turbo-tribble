package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"

	"turbo-tribble/upload"

	"github.com/gorilla/mux"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "files"
)

var psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

const MAX_UPLOAD_SIZE = 10240 * 1024 // 10MB
const FILE_ROOT_DIRECTORY = "/opt/file_manager/uploads"

func apiRoot(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, "Welcome to our API Root!")
}

func uploadFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// var uploadedFiles []file.File

	for key := range r.MultipartForm.File {

		files := r.MultipartForm.File[key]

		for _, fileHeader := range files {
			// Restrict the size of each uploaded file to 10MB.
			// To prevent the aggregate size from exceeding
			// a specified value, use the http.MaxBytesReader() method
			// before calling ParseMultipartForm()
			if fileHeader.Size > MAX_UPLOAD_SIZE {
				http.Error(w, fmt.Sprintf("The uploaded file is too big: %s. Please use a file less than 10MB in size", fileHeader.Filename), http.StatusBadRequest)
				return
			}

			// Open the file
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()

			buff := make([]byte, 512)
			_, err = file.Read(buff)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			filetype := http.DetectContentType(buff)
			if filetype != "image/jpeg" && filetype != "image/png" && filetype != "application/zip" && filetype != "application/pdf" {
				http.Error(w, "The provided file format is not allowed. Please upload files in JPEG, PNG, ZIP or PDF format", http.StatusBadRequest)
				return
			}

			_, err = file.Seek(0, io.SeekStart)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			year, month, day := time.Now().Date()
			uploadDirectory := fmt.Sprint(FILE_ROOT_DIRECTORY, "/", year, "/", month, "/", day, "/", time.Now().Hour())

			err = os.MkdirAll(uploadDirectory, os.ModePerm)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			filePath := fmt.Sprint(uploadDirectory, "/", time.Now().UnixNano())

			f, err := os.Create(fmt.Sprintf("%s%s", filePath, filepath.Ext(fileHeader.Filename)))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			defer f.Close()

			_, err = io.Copy(f, file)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", apiRoot)
	router.HandleFunc("/upload", uploadFiles).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func main() {
	sqlDB, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = upload.CreateTable(sqlDB)
	if err != nil {
		log.Fatal(err)
	}
	handleRequests()

}
