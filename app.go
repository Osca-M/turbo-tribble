package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

// sqlDB returns a pointer to our database and closes the database connection when done
func sqlDB() *sql.DB {
	sqlDB, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	// defer sqlDB.Close()
	return sqlDB
}

func apiRoot(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, "Welcome to our API Root!")
}

func uploadFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := make(map[string]string)
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		response["detail"] = "Method not allowed"
		jsonResp, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response["detail"] = fmt.Sprintf(err.Error())
		jsonResp, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
		return
	}

	// var uploadedFiles []file.File

	// Create a new context, and begin a transaction
	ctx := context.Background()
	tx, err := sqlDB().BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	// `tx` is an instance of `*sql.Tx` through which we can execute our queries

	for key := range r.MultipartForm.File {

		files := r.MultipartForm.File[key]

		for _, fileHeader := range files {
			// Restrict the size of each uploaded file to 10MB.
			// To prevent the aggregate size from exceeding
			// a specified value, use the http.MaxBytesReader() method
			// before calling ParseMultipartForm()
			if fileHeader.Size > MAX_UPLOAD_SIZE {
				w.WriteHeader(http.StatusBadRequest)
				response["detail"] = fmt.Sprintf("The uploaded file is too big: %s. Please use a file less than 10MB in size", fileHeader.Filename)
				jsonResp, err := json.Marshal(response)
				if err != nil {
					log.Printf("Error happened in JSON marshal. Err: %s", err)
				}
				w.Write(jsonResp)
				return
			}

			// Open the file
			file, err := fileHeader.Open()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				response["detail"] = fmt.Sprintf("File is damaged: %v.", err.Error())
				jsonResp, err := json.Marshal(response)
				if err != nil {
					log.Printf("Error happened in JSON marshal. Err: %s", err)
				}
				w.Write(jsonResp)
				return
			}
			defer file.Close()

			buff := make([]byte, 512)
			_, err = file.Read(buff)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				response["detail"] = "Invalid file"
				jsonResp, err := json.Marshal(response)
				if err != nil {
					log.Printf("Error happened in reading file. Err: %s", err)
				}
				w.Write(jsonResp)
				return
			}

			filetype := http.DetectContentType(buff)
			if filetype != "image/jpeg" && filetype != "image/png" && filetype != "application/zip" && filetype != "application/pdf" {
				w.WriteHeader(http.StatusBadRequest)
				response["detail"] = "The provided file format is not allowed. Please upload files in JPEG, PNG, ZIP or PDF format"
				jsonResp, err := json.Marshal(response)
				if err != nil {
					log.Printf("Error happened in file format checking. Err: %s", err)
				}
				w.Write(jsonResp)
				return
			}

			_, err = file.Seek(0, io.SeekStart)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				response["detail"] = "Invalid file"
				jsonResp, err := json.Marshal(response)
				if err != nil {
					log.Printf("Error happened in reading file. Err: %s", err)
				}
				w.Write(jsonResp)
				return
			}

			year, month, day := time.Now().Date()
			timeStamp := fmt.Sprint("/", year, "/", month, "/", day, "/", time.Now().Hour())
			uploadDirectory := fmt.Sprint(FILE_ROOT_DIRECTORY, timeStamp)

			err = os.MkdirAll(uploadDirectory, os.ModePerm)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				response["detail"] = "We are working to resolve our API outage"
				jsonResp, err := json.Marshal(response)
				if err != nil {
					log.Printf("Error happened in copying file. Err: %s", err)
				}
				w.Write(jsonResp)
				return
			}

			newFileName := time.Now().UnixNano()
			ext := filepath.Ext(fileHeader.Filename)

			filePath := fmt.Sprint(uploadDirectory, "/", newFileName, ext)

			f, err := os.Create(filePath)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				response["detail"] = "We are working to resolve our API outage"
				jsonResp, err := json.Marshal(response)
				if err != nil {
					log.Printf("Error happened in copying file. Err: %s", err)
				}
				w.Write(jsonResp)
				return
			}

			defer f.Close()

			_, err = io.Copy(f, file)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				response["detail"] = "We are working to resolve our API outage"
				jsonResp, err := json.Marshal(response)
				if err != nil {
					log.Printf("Error happened in copying file. Err: %s", err)
				}
				w.Write(jsonResp)
				return
			}

			// Save to DB
			// TODO Replace hard-coded user, handle errors better, make transactions atomic, return messages in a friendly manner
			err = upload.CreateFile(tx, ctx, key, fmt.Sprint(timeStamp, "/", newFileName, ext), "Hard-coded user")
			if err != nil {
				// Incase we find any error in the query execution, rollback the transaction
				tx.Rollback()
				w.WriteHeader(http.StatusBadRequest)
				response["detail"] = "We are working to resolve our API outage"
				jsonResp, err := json.Marshal(response)
				if err != nil {
					log.Printf("Error happened in saving to DB. Err: %s", err)
				}
				w.Write(jsonResp)
				return
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response["detail"] = "We are working to resolve our API outage"
		jsonResp, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error happened in committing to DB. Err: %s", err)
		}
		w.Write(jsonResp)
		return
	}
	w.WriteHeader(http.StatusCreated)
	response["detail"] = "Upload was successful"
	jsonResp, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error happened in responding to user. Err: %s", err)
	}
	w.Write(jsonResp)
	return

	// resp["message"] = "Upload was successful"
	// jsonResp, err := json.Marshal(resp)
	// if err != nil {
	// 	log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	// }
	// w.Write(jsonResp)
	// return
}

func getFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := make(map[string]string)
	key := mux.Vars(r)["id"]
	id, err := strconv.Atoi(key)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response["detail"] = "Invalid file ID"
		jsonResp, err := json.Marshal(response)
		if err != nil {
			log.Printf("error converting string to int")
		}
		w.Write(jsonResp)
		return
	}
	
	row := upload.GetFile(sqlDB(), id)
	var f upload.File
	err = row.Scan(&f.Id, &f.Name, &f.Path, &f.UploadedBy, &f.DateUploaded)
	if err != nil {
		log.Println("got an error")
		w.WriteHeader(http.StatusNotFound)
		response["detail"] = "File does not exist"
		jsonResp, err := json.Marshal(response)
		if err != nil {
			log.Printf("File does not exist")
		}
		w.Write(jsonResp)
		return
	}
	w.WriteHeader(http.StatusOK)
	jsonResp, err := json.Marshal(f)
	if err != nil {
		log.Printf("Error encoding File to JSON")
	}
	w.Write(jsonResp)
	return
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", apiRoot)
	router.HandleFunc("/upload", uploadFiles).Methods("POST")
	router.HandleFunc("/get-file/{id}", getFile).Methods("GET")
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
