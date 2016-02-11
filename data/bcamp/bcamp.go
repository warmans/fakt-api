package bcamp
import (
	"net/http"
	"net/url"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"github.com/texttheater/golang-levenshtein/levenshtein"
	"unicode"
	"sort"
)

type Results []*Result
func (a Results) Len() int           { return len(a) }
func (a Results) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Results) Less(i, j int) bool { return a[i].Score < a[j].Score }

type Result struct {
	Name     string
	Location string
	URL      string
	Genre    string
	Tags     []string
	Score    int
}

type Bandcamp struct {
	HTTP *http.Client
}

func (b *Bandcamp) Search(name string, location string) (Results, error) {
	searchPage, err := b.HTTP.Get("https://bandcamp.com/search?q=" + url.QueryEscape(name + " " + location))
	if err != nil {
		return nil, err
	}
	defer searchPage.Body.Close()

	doc, err := goquery.NewDocumentFromReader(searchPage.Body)
	if err != nil {
		return nil, err
	}

	//select the main data column and handle all the sub-tables
	results := make(Results, 0)
	doc.Find("#pgBd > div.search > div.leftcol > div > ul > .band").Each(func(i int, bandDiv *goquery.Selection) {
		result := b.htmlToResult(i, bandDiv)
		b.scoreResult(name, location, result)
		results = append(results, result)
	})
	sort.Sort(results)
	return results, nil
}

func (b *Bandcamp) htmlToResult(i int, bandDiv *goquery.Selection) *Result {
	result := &Result{Tags: make([]string, 0), Score: i}
	result.Name = strings.TrimSpace(bandDiv.Find(".heading").First().Text())
	result.Location = strings.TrimSpace(bandDiv.Find(".subhead").First().Text())
	result.URL = strings.TrimSpace(bandDiv.Find(".itemurl").First().Text())
	result.Genre = strings.TrimPrefix(strings.TrimSpace(bandDiv.Find(".genre").First().Text()), "genre:")
	for _, tag := range strings.Split(strings.TrimPrefix(strings.TrimSpace(bandDiv.Find(".tags").First().Text()), "tags:"), ",") {
		result.Tags = append(result.Tags, strings.TrimSpace(tag))
	}
	return result
}

func (b *Bandcamp) scoreResult(searchedName, searchedLocation string, result *Result) {
	opts := levenshtein.DefaultOptions
	opts.Matches = func(sourceCharacter rune, targetCharacter rune) bool {
		return unicode.ToLower(sourceCharacter) == unicode.ToLower(targetCharacter)
	}
	result.Score += levenshtein.DistanceForStrings([]rune(searchedName), []rune(result.Name), opts)
}
