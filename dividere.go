package main

import (
  "net/http"
  "html/template"
  "fmt"
  "io"
  "io/ioutil"
)

var uploadTemplate, _ = template.ParseFiles("upload.html")
var errorTemplate, _ = template.ParseFiles("error.html")

var listen_address = "localhost:8080"

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
  f, _, err := r.FormFile("uploaded_file")
  check(err)
  defer f.Close()
  t, err := ioutil.TempFile("./", "image-")
  check(err)
  defer t.Close()
  _, err = io.Copy(t, f)
  check(err)
  http.Redirect(w, r, "/view?id="+t.Name()[6:], 302)
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
