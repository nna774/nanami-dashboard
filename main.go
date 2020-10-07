package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/akrylysov/algnhsa"
)

var (
	PokeLocAPI = os.Getenv("POKELOC")
	PokeLocID  = os.Getenv("POKELOCID")
	POKELOC    = PokeLocAPI + "?id=" + PokeLocID

	pokelocCache *PokeLocResult

	co2API = os.Getenv("MOMOCHI")

	Tokyo, _ = time.LoadLocation("Asia/Tokyo")
	HHMM     = "15:04"
)

type Color = string

const (
	Red   Color = "red"
	Black Color = "black"

	ForDashboard = iota
	ForBlack
	ForRed
)

type Number struct {
	Approaching string   `json:"approaching"`
	Number      string   `json:"number"`
	Destination string   `json:"destination"`
	Statuses    []string `json:"statuses"`
}
type PokeLocResult struct {
	Station string   `json:"station"`
	GotAt   string   `json:"got_at"`
	Numbers []Number `json:"numbers"`
}

type Co2 struct {
	Time int64 `json:"time"`
	PPM  int   `json:"co2"`
}

type TemplateCo2 struct {
	PPM   int
	GotAt string
}

type TemplateNumber struct {
	Status0 Color
	Status1 Color
	Status2 Color
}
type TemplateValue struct {
	Co2     TemplateCo2
	Number0 TemplateNumber
	Number1 TemplateNumber
}

func fetch() (*PokeLocResult, error) {
	resp, err := http.Get(POKELOC)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result PokeLocResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	return &result, err
}

func showError(w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s: %v", msg, err)
}

func makeTemplateValue(For int, co2 *Co2, pokeLoc *PokeLocResult) *TemplateValue {
	return &TemplateValue{
		Co2: TemplateCo2{
			PPM:   co2.PPM,
			GotAt: time.Unix(co2.Time, 0).In(Tokyo).Format(HHMM),
		},
		Number1: TemplateNumber{Status0: Red, Status1: Red, Status2: Red},
	}
}

func fetchCo2() (*Co2, error) {
	resp, err := http.Get(co2API)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result Co2
	err = json.NewDecoder(resp.Body).Decode(&result)
	return &result, err
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	result, err := fetch()
	if err != nil {
		showError(w, "fetch fail", err)
		return
	}
	pokelocCache = result
	co2Result, err := fetchCo2()
	if err != nil {
		showError(w, "co2 fetch fail", err)
		return
	}
	t, err := template.ParseFiles("templates/test.svg")
	if err != nil {
		showError(w, "template fail", err)
		return
	}

	v := makeTemplateValue(ForDashboard, co2Result, pokelocCache)
	err = t.Execute(w, v)
	if err != nil {
		fmt.Fprintf(w, "template execute fail: %v", err)
	}

}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		result, err := fetch()
		fmt.Fprint(w, result, err)
		pokelocCache = result
	})
	if os.Getenv("NANAMI_ENV") == "development" {
		panic(http.ListenAndServe(":8000", nil))
	} else {
		algnhsa.ListenAndServe(nil, &algnhsa.Options{RequestType: algnhsa.RequestTypeAPIGateway})
	}
}
