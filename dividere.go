package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	pfp "path/filepath"
)

var uploadTemplate, _ = template.ParseFiles("html/upload.html")
var errorTemplate, _ = template.ParseFiles("html/error.html")

var listen_address = "localhost:8080"
var filesDirectory = "./f"
var max_upload_size int64 = 100000

func checkAndCreateDir(path string) {
	absPath, err := pfp.Abs(path)
	if err != nil {
		log.Fatal("Cannot determine absolute path for file storage: " +
			err.Error())
	}
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("File directory %s does not exist - creating it\n", absPath)
			if err := os.Mkdir(absPath, os.ModeDir|0777); err != nil {
				log.Fatal("Cannot create file directory " + absPath + ": " +
					err.Error())
			}
		} else {
			log.Fatal("Cannot use file storage directory " + filesDirectory +
				": " + err.Error())
		}
	}
}

func errorHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				w.WriteHeader(500)
				errorTemplate.Execute(w, e)
			}
		}()
		fn(w, r)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		uploadTemplate.Execute(w, nil)
		return
	}
	err := r.ParseMultipartForm(10000)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("successfully parsed multipart form")
	}
	for key, value := range r.MultipartForm.Value {
		fmt.Printf("%s:%s", key, value)
	}

	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			//file, _ := fileHeader.Open()
			//http://golang.org/pkg/io/ioutil/#TempDir
			path := fmt.Sprintf("files/%s", fileHeader.Filename)
			fmt.Println("Would save to " + path)
			//buf, _ := ioutil.ReadAll(file)
			//ioutil.WriteFile(path, buf, os.ModePerm)
		}
	}

	fmt.Printf("%v\n", r.FormValue("filename"))
	f, _, err := r.FormFile("uploaded_file")
	if err != nil {
		fmt.Println("Error retrieving file from multipart form: " +
			err.Error())
		return
	}
	defer f.Close()
	t, err := ioutil.TempFile("./f", "")
	if err != nil {
		fmt.Println("Cannot create temp file: " + err.Error())
		return
	}
	defer t.Close()
	_, err = io.Copy(t, f)
	if err != nil {
		fmt.Println("Cannot copy data to file: " + err.Error())
		return
	}
	http.Redirect(w, r, "/view?id="+t.Name()[6:], 302)
}

func view(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image")
	http.ServeFile(w, r, "image-"+r.FormValue("id"))
}

func main() {
	http.HandleFunc("/", errorHandler(upload))
	http.HandleFunc("/view", errorHandler(view))
	checkAndCreateDir(filesDirectory)
	fmt.Println("Starting server at " + listen_address)
	log.Fatal(http.ListenAndServe(listen_address, nil))
	// Simple static webserver:
	//log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir("/usr/share/doc"))))
}
