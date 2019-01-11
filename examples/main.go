// Package examples @author K·J Create at 2019-01-11 13:56
package main

import (
	"fmt"
	"log"
	"net/http"

	"adventurer"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	stories, err := adventurer.LoadStories("./stories/stories.yaml")
	if nil != err {
		log.Fatalln(err.Error())
	}
	hook := make(adventurer.StoryHook, 0)
	hook["header"] = NewHeaderHook()
	a, err := adventurer.NewAdventurer(Demo{}, stories,
		adventurer.NewProfile("/demo/about", "0.0.1", "1.11.2", "", "test"),
		&hook)
	if nil != err {
		log.Fatalln(err.Error())
	}
	a.SetCros(true)
	http.HandleFunc("/", a.Explore)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 2111), nil))
}

// Demo demo
type Demo struct {
}

// DemoHandler demo handler
func (d *Demo) DemoHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}

type HeaderHook struct {
}

func NewHeaderHook() adventurer.Hook {
	return &HeaderHook{}
}

func (h *HeaderHook) Fire(prerequisite []string, equipment adventurer.Equipment) (bool, error) {
	if nil == prerequisite || len(prerequisite) < 1 {
		return false, nil
	}
	if "true" == prerequisite[0] {
		if "" == equipment.Header.Get("device") {
			return false, nil
		}
	}
	return true, nil
}
