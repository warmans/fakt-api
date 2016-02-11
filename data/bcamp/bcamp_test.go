package bcamp

import (
"net/http"
"testing"
	"fmt"
)

func TestSearch(t *testing.T) {
	bcamp := Bandcamp{HTTP: http.DefaultClient}
	result, err := bcamp.Search("Is Dodelijk", "Nürnberg")
	if err !=nil {
		fmt.Print(err.Error())
		return
	}
	for _, res := range result {
		fmt.Printf("%+v", res)
	}
}
