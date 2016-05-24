package main

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	set "github.com/deckarep/golang-set"
	"github.com/djimenez/iconv-go"
	"github.com/shogo82148/go-mecab"
)

const targetURL1 string = "http://d.hatena.ne.jp/naoya/20160204"
const targetURL2 string = "http://d.hatena.ne.jp/naoya/20151026"

func main() {
	urlWordMap := map[string][]string{}
	targetURLs := []string{targetURL1, targetURL2}
	for _, url := range targetURLs {
		// Load the URL
		res, err := http.Get(url)
		encoding := "utf-8"
		cType := res.Header.Get("Content-Type")
		if strings.Contains(cType, "charset=") {
			index := strings.Index(cType, "=")
			encoding = cType[index+1:]
		}
		defer res.Body.Close()

		// Convert the designated charset HTML to utf-8 encoded HTML.
		// `charset` being one of the charsets known by the iconv package.
		utfBody, err := iconv.NewReader(res.Body, encoding, "utf-8")
		if err != nil {
			// handler error
		}

		// use utfBody using goquery
		doc, err := goquery.NewDocumentFromReader(utfBody)
		if err != nil {
			fmt.Print("url scarapping failed")
		}
		doc.Find("script").Remove()
		doc.Find("noscript").Remove()
		doc.Find("style").Remove()
		query := doc.Find("body").Text()

		if err != nil {
			panic("failed to remove tags")
		}
		tagger, err := mecab.New(map[string]string{})
		if err != nil {
			panic(err)
		}
		defer tagger.Destroy()
		tagger.Parse("")

		node, err := tagger.ParseToNode(query)
		if err != nil {
			panic(err)
		}
		var words []string
		for ; node != (mecab.Node{}); node = node.Next() {
			features := strings.Split(node.Feature(), ",")
			if features[0] == "名詞" {
				words = append(words, node.Surface())
			}
		}
		urlWordMap[url] = words
	}
	uniqueWords := set.NewSet()
	for _, words := range urlWordMap {
		for _, word := range words {
			uniqueWords.Add(word)
		}
	}
	flags := [][]int{}
	for _, words := range urlWordMap {
		flags = append(flags, MakeFlag(ToStringSlice(uniqueWords.ToSlice()), words))
	}
	innerProd := 0.0
	for i, _ := range flags[0] {
		innerProd = innerProd + float64((flags[0][i] * flags[1][i]))
	}
	result := innerProd / (Norm(flags[0]) * Norm(flags[1]))
	fmt.Println(result)

}

func ToStringSlice(a []interface{}) []string {
	var retData []string
	for _, val := range a {
		retData = append(retData, val.(string))
	}
	return retData
}

func Norm(a []int) float64 {
	sum := 0.0
	for _, val := range a {
		sum = sum + math.Abs(float64(val*val))
	}
	return sum
}

func MakeFlag(unique []string, words []string) []int {
	var flags []int
	for _, uniqueWord := range unique {
		if Contains(words, uniqueWord) {
			flags = append(flags, 1)
		} else {
			flags = append(flags, 0)
		}
	}
	return flags
}

func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}
