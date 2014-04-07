package main

import (
  "net/http"
  "html/template"
  "fmt"
  "os"
  "io/ioutil"
)

var uploadTemplate, _ = template.ParseFiles("upload.html")
var errorTemplate, _ = template.ParseFiles("error.html")

var listen_address = "localhost:8080"
var max_upload_size int64 = 100000

func check(err error) { if err != nil { panic(err) } }

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
  err := r.ParseMultipartForm(max_upload_size);
  if err != nil {
    fmt.Println(err.Error())
    return
  } 
  for key, value := range r.MultipartForm.Value {
    fmt.Printf("%s:%s\n", key, value)
  }

  for _, fileHeaders := range r.MultipartForm.File {
    for _, fileHeader := range fileHeaders {
      file, _ := fileHeader.Open()
      path := fmt.Sprintf("files/%s", fileHeader.Filename)
      fmt.Println("Saving to " + path);
      buf, _ := ioutil.ReadAll(file)
      ioutil.WriteFile(path, buf, os.ModePerm)
    }
  }

  //fmt.Printf("%v\n", r.FormValue("filename"))
  //f, _, err := r.FormFile("uploaded_file")
  //if err != nil {
  //  fmt.Println("Error retrieving file from multipart form: " +
  //  err.Error());
  //  return
  //} 
  //defer f.Close()
  //t, err := ioutil.TempFile("./f", "")
  //if err != nil {
  //  fmt.Println("Cannot create temp file: " + err.Error());
  //  return
  //} 
  //defer t.Close()
  //_, err = io.Copy(t, f)
  //check(err)
  http.Redirect(w, r, "/view?id="+"foobar"[6:], 302)
}

func view(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "image")
  http.ServeFile(w, r, "image-"+r.FormValue("id"))
}

func main() {
  http.HandleFunc("/", errorHandler(upload))
  http.HandleFunc("/view", errorHandler(view))
  fmt.Println("Starting server at " + listen_address)
  http.ListenAndServe(listen_address, nil)
  // Simple static webserver:
  //log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir("/usr/share/doc"))))
}
