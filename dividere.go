package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	pfp "path/filepath"
  "github.com/dchest/uniuri"
  "code.google.com/p/gcfg"
)

type ConfigurationData struct {
    Network struct {
        Host string
        Port int32
    }
    Storage struct {
      Directory string
      MaxFileSize int64
    }
}

var Cfg ConfigurationData

var uploadTemplate, _ = template.ParseFiles("html/upload.html")
var errorTemplate, _ = template.ParseFiles("html/error.html")
var showTemplate, _ = template.ParseFiles("html/show.html")


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
			log.Fatal("Cannot use file storage directory " + absPath +
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
	err := r.ParseMultipartForm(Cfg.Storage.MaxFileSize)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("successfully parsed multipart form")
	}
  randomURI := uniuri.New()
  fmt.Printf("URI: %s\n", randomURI)
  // Create a temp directory
  absPath, err := pfp.Abs(Cfg.Storage.Directory)
  if err != nil {
    log.Fatal("Cannot determine absolute path for file storage: " +
    err.Error())
  }
  tmpDir := pfp.Join(absPath, randomURI)
  err = os.Mkdir(tmpDir, os.ModePerm)
  if err != nil {
    log.Fatal("Cannot create temp dir for file storage: " + err.Error())
  }
  for _, fileHeaders := range r.MultipartForm.File {
    for _, fileHeader := range fileHeaders {
      file, _ := fileHeader.Open()

      // Calculate path for file storage
      path := fmt.Sprintf("%s/%s", tmpDir, fileHeader.Filename)
      fmt.Printf("Saving %s\n", path)
      buf, _ := ioutil.ReadAll(file)
      ioutil.WriteFile(path, buf, os.ModePerm)
    }
  }
  http.Redirect(w, r, "/show?id="+randomURI, 302)
}

func view(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "image")
  http.ServeFile(w, r, "image-"+r.FormValue("id"))
}

func show(w http.ResponseWriter, r *http.Request) {
  type Data struct {
    Directory string
    FilesURI []string
  }
  id := r.FormValue("id");
  //d := Data{Directory: id, FilesURI: []string{"foo", "bar", "baz"} }
  d := Data{Directory: id, FilesURI: []string{}}
// Create a temp directory
  absPath, err := pfp.Abs(Cfg.Storage.Directory)
  if err != nil {
    log.Fatal("Cannot determine absolute path for file storage: " +
    err.Error())
  }
  tmpDir := pfp.Join(absPath, id)
  // TODO: This is vulnerable to a directory traversal attack.
  // Countermeasures needed!
  files, _ := ioutil.ReadDir(tmpDir);
  for _, f:= range files {
    d.FilesURI = append(d.FilesURI, f.Name())
  }
  err = showTemplate.Execute(w, d)
  if err != nil {
    fmt.Printf("Error executing template: %s", err.Error())
  }
}

func main() {
  err := gcfg.ReadFileInto(&Cfg, "dividere.conf")
  if err != nil {
    log.Fatal("Cannot read configuration file: " + err.Error());
  }
  listenAddress := fmt.Sprintf("%s:%d", Cfg.Network.Host, Cfg.Network.Port);
  checkAndCreateDir(Cfg.Storage.Directory)
  http.HandleFunc("/", errorHandler(upload))
  http.HandleFunc("/view", errorHandler(view))
  http.HandleFunc("/show", errorHandler(show))

  fmt.Println("Starting server at " + listenAddress)
  log.Fatal(http.ListenAndServe(listenAddress, nil))
  // Simple static webserver:
  //log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir("/usr/share/doc"))))
}
