package main

import (
  "html/template"
  "io/ioutil"
  "net/http"
  "regexp"
)

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type Page struct {
  Title string
  Body []byte
  Display template.HTML
}

func (p *Page) save() error {
  filename := p.Title + ".txt"
  return ioutil.WriteFile("data/"+filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
  filename := title + ".txt"
  body, err := ioutil.ReadFile("data/"+filename)
  if err != nil {
    return nil, err
  }
  return &Page{Title: title, Body: body}, nil
}

var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page){
  err := templates.ExecuteTemplate(w, tmpl + ".html", p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
      http.NotFound(w, r)
      return
    }
    fn(w, r, m[2])
  }
}

var linkRegexp = regexp.MustCompile("\\[([a-zA-Z0-9]+)\\]")

func viewHandler(w http.ResponseWriter, r *http.Request, title string){
  p, err := loadPage(title)
  if err != nil {
    http.Redirect(w, r, "/edit/" + title, http.StatusFound)
    return
  }

  escapedBody := []byte(template.HTMLEscapeString(string(p.Body)))

  p.Display = template.HTML(linkRegexp.ReplaceAllFunc(escapedBody, func(str []byte) []byte {
    match := linkRegexp.FindStringSubmatch(string(str))
    out := []byte("<a href=\"/view/" + match[1] + "\">" + match[1] + "</a>")
    return out
    }))
  renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string){
  p, err := loadPage(title)
  if err != nil {
    p = &Page{Title: title}
  }
  renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string){
  body := r.FormValue("body")
  p := &Page{Title: title, Body: []byte(body)}
  err := p.save()
  if err != nil {
    http.Redirect(w, r, "/edit/" + title, http.StatusFound)
    return
  }
  http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func frontPageHandler(w http.ResponseWriter, r *http.Request){
  http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func main() {
  http.HandleFunc("/view/", makeHandler(viewHandler))
  http.HandleFunc("/edit/", makeHandler(editHandler))
  http.HandleFunc("/save/", makeHandler(saveHandler))
  http.HandleFunc("/", frontPageHandler)
  http.ListenAndServe(":8080", nil)
}
