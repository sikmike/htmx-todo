package todohttp

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/quii/todo/domain/todo"
)

var (
	//go:embed "templates/*"
	todoTemplates embed.FS

	//go:embed static
	static embed.FS
)

type TodoHandler struct {
	http.Handler

	list  *todo.List
	templ *template.Template
}

func NewTodoHandler(service *todo.List) *TodoHandler {
	router := mux.NewRouter()
	handler := &TodoHandler{
		Handler: router,
		list:    service,
	}

	router.HandleFunc("/", handler.index).Methods(http.MethodGet)

	router.HandleFunc("/todos", handler.add).Methods(http.MethodPost)
	router.HandleFunc("/todos", handler.search).Methods(http.MethodGet)
	router.HandleFunc("/todos/{ID}/toggle", handler.toggle).Methods(http.MethodPost)
	router.HandleFunc("/todos/{ID}", handler.delete).Methods(http.MethodDelete)
	router.HandleFunc("/todos/sort", handler.reOrder).Methods(http.MethodPost)

	staticHandler, err := newStaticHandler()
	if err != nil {
		panic(err)
	}
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticHandler))

	templ, err := template.ParseFS(todoTemplates, "templates/*.gohtml")
	if err != nil {
		panic(err)
	}

	handler.templ = templ

	return handler
}

func (t *TodoHandler) index(w http.ResponseWriter, r *http.Request) {
	if err := t.templ.ExecuteTemplate(w, "index.gohtml", t.list.Todos()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (t *TodoHandler) add(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	t.list.Add(r.FormValue("description"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (t *TodoHandler) toggle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["ID"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	t.templ.ExecuteTemplate(w, "item.gohtml", t.list.ToggleDone(id))
}

func (t *TodoHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["ID"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	t.list.Delete(id)
}

func (t *TodoHandler) reOrder(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	t.list.ReOrder(r.Form["id"])
	if err := t.templ.ExecuteTemplate(w, "items.gohtml", t.list.Todos()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (t *TodoHandler) search(w http.ResponseWriter, r *http.Request) {
	if err := t.templ.ExecuteTemplate(w, "items.gohtml", t.list.Search(r.URL.Query().Get("search"))); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func newStaticHandler() (http.Handler, error) {
	lol, err := fs.Sub(static, "static")
	if err != nil {
		return nil, err
	}
	return http.FileServer(http.FS(lol)), nil
}
