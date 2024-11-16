package probe

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type ApiHandler struct {
	urlsChan chan []string
	urlChan  chan string
	urlsPtr  *[]string
}

func GetApiHandler(urls *[]string, urlsUpdateChan chan []string, urlAddChan chan string) *ApiHandler {
	return &ApiHandler{
		urlsChan: urlsUpdateChan,
		urlChan:  urlAddChan,
		urlsPtr:  urls,
	}
}

func (a *ApiHandler) Router() *mux.Router {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", a.homePage)
	myRouter.HandleFunc("/urls/update", a.updateUrls)
	myRouter.HandleFunc("/urls/add", a.addUrl)

	return myRouter
}

func (a *ApiHandler) homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!\n")
	fmt.Fprintf(w, "Current list: %v", a.urlsPtr)
}

// ToDo: validate urls
func (a *ApiHandler) updateUrls(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	us := q["urls"]
	log.Debug().Interface("urls", us).Msg("Received load request")
	a.urlsChan <- us
	fmt.Fprintf(w, "Hello %v", us)
}

func (a *ApiHandler) addUrl(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	url := q.Get("url")
	a.urlChan <- url
	fmt.Fprintf(w, "Hello %v", url)
}
