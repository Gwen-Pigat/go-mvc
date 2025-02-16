package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sandbox/initializers"
	"sandbox/models"
	"text/template"
	"time"

	"github.com/joho/godotenv"
)

const mainPath string = "/"
const viewsDir string = "views/"
const staticDir string = "static"

type Wrapper struct {
	writer  http.ResponseWriter
	request *http.Request
}

var wrapper *Wrapper
var db *sql.DB

func init() {
	godotenv.Load()
	db, err := initializers.ConnectDB()
	if err != nil {
		log.Fatal("Erreur init", err, db)
	}
}

func main() {
	defer db.Close()
	fs := http.FileServer(http.Dir(staticDir))
	GET(mainPath, Index)
	http.Handle("/"+staticDir+"/", http.StripPrefix("/"+staticDir+"/", fs))
	POST("/clicked", TriggerAction)
	GET("/items", ItemsList)
	GET("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, staticDir+"/favicon.png")
	})
	port := "3000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Error server", err)
	}
	fmt.Println("Hello world server")
}

func GET(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	fmt.Printf("Wrapper GET %v", path)
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			wrapper.Error(&models.Pages{
				Title:   "Wrong method",
				Content: "Method not autorized",
			}, 405)
			return
		}
		handler(w, r)
	})
}

func POST(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	fmt.Printf("Wrapper POST %v", path)
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			wrapper.Error(&models.Pages{
				Title:   "Wrong method",
				Content: "Method not autorized",
			}, 405)
			return
		}
		handler(w, r)
	})
}

func ItemsList(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM items WHERE is_active=1 ORDER BY id DESC")
	if err != nil {
		wrapper.Error(&models.Pages{
			Title:   "No results found",
			Content: "No results found",
		})
		return
	}
	defer rows.Close()
	type Result struct {
		ID      int
		Value   string
		DateAdd time.Time
	}
	results := make(map[int]Result)
	for rows.Next() {
		var result Result
		if err := rows.Scan(&result.ID, &result.Value, &result.DateAdd); err != nil {
			wrapper.Error(&models.Pages{
				Title:   "Error fetch items",
				Content: err.Error(),
			}, 500)
			return
		}
		results[result.ID] = result
	}
	wrapper.Render(&models.Pages{
		Title:   "Items",
		Content: "",
		Data:    results,
	}, "items")
}

func TriggerAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method is not allowed"))
		return
	}
	tpl, err := template.ParseFiles(
		viewsDir + "ajax/clicked.ajax.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Template cannot be found => " + err.Error()))
		return
	}
	i := 1
	data := map[int]int{}
	for i < 10 {
		data[i] = i
		i++
	}
	if err := tpl.Execute(w, map[string]interface{}{
		"data": data,
	}); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Template cannot be found => " + err.Error()))
		return
	}
	w.Write([]byte("Action is replaced"))
}

func (wrapper Wrapper) Render(content *models.Pages, view string, status ...int) {
	tpl, err := template.ParseFiles(
		viewsDir+"base.html",
		viewsDir+"includes/navbar.include.html",
		viewsDir+"includes/foot.include.html",
		viewsDir+"includes/head.include.html",
		viewsDir+view+".html",
	)
	if err != nil {
		http.Error(wrapper.writer, "Unabled to load page", http.StatusInternalServerError)
		log.Fatal(err)
	}
	code := http.StatusNotFound
	if len(status) > 0 {
		code = status[0]
	}
	content.LoadDefaultScripts()
	wrapper.writer.WriteHeader(code)
	if err := tpl.Execute(wrapper.writer, content); err != nil {
		http.Error(wrapper.writer, "Unabled to load template", http.StatusInternalServerError)
		log.Fatal(err, tpl)
	}
}

func (wrapper Wrapper) Error(result *models.Pages, code ...int) {
	fmt.Println("Error layout", wrapper.request.URL.Path)
	statusCode := 404
	if len(code) > 0 {
		statusCode = code[0]
	}
	if result == nil {
		result = &models.Pages{
			Title:   "Page not found",
			Content: "Page cannot be found",
		}
	}
	wrapper.Render(result, "error", statusCode)
}

func Index(w http.ResponseWriter, r *http.Request) {
	if wrapper.request.URL.Path != mainPath {
		wrapper.Error(nil)
		return
	}
	wrapper.Render(&models.Pages{
		Title:   "Page d'accueil",
		Content: "Contenu de ma page d'accueil",
	}, "index", http.StatusOK)
	fmt.Println("Index is served", wrapper.request.URL.Path)
}
