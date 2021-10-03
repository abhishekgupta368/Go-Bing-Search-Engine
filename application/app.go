package application

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/SearchEngine/model"
)

var userAgents = []string{
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/603.3.8 (KHTML, like Gecko) Version/10.1.2 Safari/603.3.8",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_2 like Mac OS X) AppleWebKit/603.2.4 (KHTML, like Gecko) Version/10.0 Mobile/14F89 Safari/602.1",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_2 like Mac OS X) AppleWebKit/603.2.4 (KHTML, like Gecko) FxiOS/8.1.1b4948 Mobile/14F89 Safari/603.2.4",
	"Mozilla/5.0 (iPad; CPU OS 10_3_2 like Mac OS X) AppleWebKit/603.2.4 (KHTML, like Gecko) Version/10.0 Mobile/14F89 Safari/602.1",
	"Mozilla/5.0 (Linux; Android 4.3; GT-I9300 Build/JSS15J) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.125 Mobile Safari/537.36",
	"Mozilla/5.0 (Android 4.3; Mobile; rv:54.0) Gecko/54.0 Firefox/54.0",
	"Mozilla/5.0 (Linux; Android 4.3; GT-I9300 Build/JSS15J) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.91 Mobile Safari/537.36 OPR/42.9.2246.119956",
	"Opera/9.80 (Android; Opera Mini/28.0.2254/66.318; U; en) Presto/2.12.423 Version/12.16",
}

var bingDomains = map[string]string{
	"com": "",
}

func RandomUserAgent() string {
	return userAgents[1]
}

func FirstParameter(number, count int) int {
	if number == 0 {
		return number + 1
	}
	return number*count + 1
}

func BuildBingUrls(searchTerm, country string, pages, count int) ([]string, error) {
	toScrape := []string{}
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm, " ", "+", -1)
	if countryCode, found := bingDomains[country]; found {
		for i := 0; i < pages; i++ {
			first := FirstParameter(i, count)
			scrapeUrl := fmt.Sprintf("https://www.bing.com/search?q=%s&first=%d$count%d%s", searchTerm, first, count, countryCode)
			toScrape = append(toScrape, scrapeUrl)
		}
	} else {

		return nil, fmt.Errorf("count(%s) not found in domain database", countryCode)
	}
	return toScrape, nil
}

func GetScrapeClient(proxyString interface{}) *http.Client {
	switch v := proxyString.(type) {
	case string:
		{
			proxyUrl, _ := url.Parse(v)
			return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
		}
	default:
		{
			return &http.Client{}
		}
	}
}

func ScrapeClientRequest(searchUrl string, proxyString interface{}) (*http.Response, error) {
	baseClient := GetScrapeClient(proxyString)
	req, _ := http.NewRequest("GET", searchUrl, nil)
	req.Header.Add("User-Agent", RandomUserAgent())
	res, err := baseClient.Do(req)

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("200 Status not found")
	}
	if err != nil {
		return nil, err
	}
	return res, err
}

func BingResultParser(resp *http.Response, rank int) ([]model.SearchResult, error) {
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}
	ans := []model.SearchResult{}
	sel := doc.Find("li.b_algo")
	rank++

	for i := range sel.Nodes {
		item := sel.Eq(i)
		linkTag := item.Find("a")
		link, _ := linkTag.Attr("href")
		titleTag := item.Find("h2")
		descTag := item.Find("div.b_caption p")
		desc := descTag.Text()
		title := titleTag.Text()
		link = strings.Trim(link, " ")
		if link != " " && link != "#" && !strings.HasPrefix(link, "/") {
			result := model.SearchResult{
				rank,
				link,
				title,
				desc,
			}
			ans = append(ans, result)
			rank++
		}
	}
	return ans, nil
}

func BingScrape(searchTerm, country string, proxyString interface{}, pages, count, backoff int) ([]model.SearchResult, error) {
	searchResultList := []model.SearchResult{}
	bingPages, err := BuildBingUrls(searchTerm, country, pages, count)

	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	} else {
		for _, page := range bingPages {
			rank := len(searchResultList)
			res, err := ScrapeClientRequest(page, proxyString)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			data, err := BingResultParser(res, rank)

			if err != nil {
				return nil, err
			}

			for _, result := range data {
				searchResultList = append(searchResultList, result)
			}
			time.Sleep(time.Duration(backoff) * time.Second)
		}
	}

	return searchResultList, nil
}

// func ResultParser() string {

// }

func StartApplication() {
	var searchQuery string

	fmt.Println("Enter your search query")
	fmt.Scanln(&searchQuery)

	res, err := BingScrape(searchQuery, "com", nil, 2, 15, 30)
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println(res)
	}
}
