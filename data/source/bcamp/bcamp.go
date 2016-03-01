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

func (a Results) Len() int { return len(a) }
func (a Results) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Results) Less(i, j int) bool { return a[i].Score < a[j].Score }

type Result struct {
	Name       string
	Location   string
	URL        string
	Genre      string
	Tags       []string
	Art        string
	Score      int
}

type ArtistPage struct {
	Bio   string
	Links []*Link
}

type Link struct {
	URI  string
	Text string
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
		result := b.processSearchResult(i, bandDiv)
		b.scoreResult(name, location, result)
		results = append(results, result)
	})
	sort.Sort(results)
	return results, nil
}

func (b *Bandcamp) processSearchResult(i int, bandDiv *goquery.Selection) *Result {

	//data from directly on the search results page
	result := &Result{Tags: make([]string, 0), Score: i}
	result.Name = strings.TrimSpace(bandDiv.Find(".heading").First().Text())
	result.Location = strings.TrimSpace(bandDiv.Find(".subhead").First().Text())
	result.URL = strings.TrimSpace(bandDiv.Find(".itemurl").First().Text())
	result.Genre = strings.TrimPrefix(strings.TrimSpace(bandDiv.Find(".genre").First().Text()), "genre:")
	for _, tag := range strings.Split(strings.TrimPrefix(strings.TrimSpace(bandDiv.Find(".tags").First().Text()), "tags:"), ",") {
		result.Tags = append(result.Tags, strings.TrimSpace(tag))
	}
	result.Art = strings.TrimSpace(bandDiv.Find(".artcont .art img").First().AttrOr("src", ""))
	return result
}

func (b *Bandcamp) GetArtistPageInfo(artistURL string) *ArtistPage {

	a := &ArtistPage{Bio: "", Links: make([]*Link, 0)}
	if artistURL == "" {
		return a
	}

	searchPage, err := b.HTTP.Get(artistURL)
	if err != nil {
		return a
	}
	defer searchPage.Body.Close()
	doc, err := goquery.NewDocumentFromReader(searchPage.Body)
	if err != nil {
		return a
	}

	doc.Find("#bio-container").Each(func(i int, bioContainer *goquery.Selection) {
		a.Bio = strings.TrimSpace(bioContainer.Find(".signed-out-artists-bio-text meta").First().AttrOr("content", ""))
	})
	doc.Find("#band-links li a").Each(func(i int, atag *goquery.Selection) {
		a.Links = append(a.Links, &Link{URI: atag.AttrOr("href", ""), Text: strings.TrimSpace(atag.Text())})
	})
	return a
}

func (b *Bandcamp) scoreResult(searchedName, searchedLocation string, result *Result) {
	opts := levenshtein.DefaultOptions
	opts.Matches = func(sourceCharacter rune, targetCharacter rune) bool {
		return unicode.ToLower(sourceCharacter) == unicode.ToLower(targetCharacter)
	}
	result.Score += levenshtein.DistanceForStrings([]rune(searchedName), []rune(result.Name), opts)
}


