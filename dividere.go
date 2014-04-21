package main

import (
	"code.google.com/p/gcfg"
	"fmt"
	"github.com/dchest/uniuri"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	pfp "path/filepath"
	"strings"
)

type ConfigurationData struct {
	Network struct {
		Host        string
		Port        int32
		ExternalURL string
	}
	Storage struct {
		Directory   string
		BaseURL     string
		MaxFileSize int64
	}
	Misc struct {
		BufferSize int64
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

func mkReadableSize(byteSize int64) string {
	unit := "B"
	value := byteSize
	if byteSize > 1024 {
		unit = "KB"
		value = byteSize / 1024
	}
	if byteSize > 1024*1024 {
		unit = "MB"
		value = byteSize / (1024 * 1024)
	}
	if byteSize > 1024*1024*1024 {
		unit = "GB"
		value = byteSize / (1024 * 1024 * 1024)
	}
	return fmt.Sprintf("%d %s", value, unit)
}

func show(w http.ResponseWriter, r *http.Request) {
	type FileData struct {
		FileName         string
		FileURI          string
		FileSizeReadable string
	}
	type Data struct {
		Directory    string
		FileList     []FileData
		ErrorCode    int
		ErrorMessage string
	}
	id := r.FormValue("id")
	d := new(Data)
	d.Directory = id
	d.FileList = []FileData{}
	//d := Data{ Directory: id, FileData: []{FileName: "Foo", FileURI: "Uschi"} }
	// Create a temp directory
	absPath, err := pfp.Abs(Cfg.Storage.Directory)
	if err != nil {
		log.Fatal("Cannot determine absolute path for file storage: " +
			err.Error())
	}
	// Note: The runtime complexity of the directory traversal might be
	// too high for big installations, but let's worry about that when we
	// have a big installation.
	found := false
	_ = pfp.Walk(absPath, func(currentPath string, info os.FileInfo, err error) error {
		if info.IsDir() && info.Name() == id {
			found = true
			// for all files in our target directory: add files to our info
			// structure.
			_ = pfp.Walk(currentPath, func(currentPath string,
				info os.FileInfo, err error) error {
				if !info.IsDir() {
					components := []string{
						Cfg.Network.ExternalURL,
						Cfg.Storage.BaseURL,
						d.Directory,
						info.Name()}
					uri := strings.Join(components, "/")
					fileEntry := FileData{FileName: info.Name(), FileURI: uri,
						FileSizeReadable: mkReadableSize(info.Size())}
					d.FileList = append(d.FileList, fileEntry)
				}
				return nil
			})
		}
		return nil
	})
	if found {
		err = showTemplate.Execute(w, d)
		if err != nil {
			fmt.Printf("Error executing template: %s", err.Error())
		}
	} else {
		// the data directory was not found - render a 404 error page
		w.WriteHeader(http.StatusNotFound)
		d.ErrorCode = http.StatusNotFound
		d.ErrorMessage = "These are not the bytes you're looking for."
		err = errorTemplate.Execute(w, d)
	}
}

func main() {
	err := gcfg.ReadFileInto(&Cfg, "dividere.conf")
	if err != nil {
		log.Fatal("Cannot read configuration file: " + err.Error())
	}
	listenAddress := fmt.Sprintf("%s:%d", Cfg.Network.Host, Cfg.Network.Port)
	checkAndCreateDir(Cfg.Storage.Directory)
	http.HandleFunc("/", errorHandler(upload))
	http.HandleFunc("/show", errorHandler(show))
	http.Handle(Cfg.Storage.BaseURL, http.StripPrefix(Cfg.Storage.BaseURL,
		http.FileServer(http.Dir(Cfg.Storage.Directory))))
	fmt.Println("Starting server at " + listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
