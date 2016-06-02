package webapp

import (
	"html/template"
	"net/http"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
	"google.golang.org/appengine/datastore"
)

type Greeting struct {
	Author string
	Content string
	Date time.Time
}

var guestbookTemplate = template.Must(template.New("book").Parse(guestbookTemplateHTML))

func init() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/sign", signHandler)
	http.HandleFunc("/login", loginHandler)
}


func indexHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Greeting").Order("-Date").Limit(10)
	greetings := make([]Greeting, 0, 10)
	if _, err := q.GetAll(c, &greetings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Debugf(c, "%v", &greetings)
	if err := guestbookTemplate.Execute(w, greetings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}


func signHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	g := Greeting {
		Content: r.FormValue("content"),
		Date: time.Now(),
		}
	if u := user.Current(c); u != nil {
		g.Author = u.String()
	}
	
	_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Greeting", nil), &g)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w,r,"/", http.StatusFound)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return
	}
	
	http.Redirect(w,r,"/", http.StatusFound)
}

const guestbookTemplateHTML = `
<html>
  <body>
    {{range .}}
      {{with .Author}}
        <p><b>{{.}}</b> wrote:</p>
      {{else}}
        <p>An anonymous person wrote:</p>
      {{end}}
      <pre>{{.Content}}</pre>
    {{end}}
    <form action="/sign" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Sign Guestbook"></div>
    </form>
    <a href="/login">Login to see your username</a>
  </body>
</html>
`
