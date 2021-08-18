package main

import (
	"log"
	"math"
	"regexp"
	"strconv"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
)

func main() {
	c := colly.NewCollector()
	c.AllowURLRevisit = true
	c.AllowedDomains = []string{"www.kanplaza.com", "kanplaza.com"}
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 25,
	})
	rp, err := proxy.RoundRobinProxySwitcher(
		"http://aaa:aaa@103.85.24.11:3128",
		"http://aaa:aaa@103.85.24.12:3128",
		"http://aaa:aaa@103.85.24.13:3128",
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	c.SetProxyFunc(rp)
	// //section[@class="side_category"]/ul/li/a[@id="link"]/@href
	c.OnHTML("section.side_category", func(e *colly.HTMLElement) {
		e.ForEach("ul>li>a", func(_ int, el *colly.HTMLElement) {
			href := el.Attr("href")
			if href != "" {
				catogries_list(c, el.Request.AbsoluteURL(href))
				// log.Println(el.Request.AbsoluteURL(href))
			}
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("TOP: ", r.Request.URL.String(), "failed with: ", err)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visit form: ", r.URL.String())
	})

	c.Visit("https://kanplaza.com/ec/cmShopTopPage4.html")
}

func catogries_list(webColly *colly.Collector, url string) {
	categoryColly := webColly.Clone()
	categoryColly.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 16,
	})

	categoryColly.OnHTML("p.total_count", func(e *colly.HTMLElement) {
		reg := regexp.MustCompile(`\D*`)
		pageCount := reg.ReplaceAllString(e.Text, "")
		pageNum, _ := strconv.ParseFloat(pageCount, 64)
		page := int(math.Ceil(pageNum / 80.0))
		url := e.Request.URL.String()
		for i := 0; i < page; i++ {
			new_url := url + "&page=" + strconv.Itoa(page)
			categoryNext(webColly, new_url)
		}
	})
	categoryColly.OnError(func(resp *colly.Response, err error) {
		log.Println("Category List: ", resp.Request.URL.String(), "failed with response: ", err)
	})

	categoryColly.Visit(url + "&dnumber=80")
}

func categoryNext(webColly *colly.Collector, url string) {
	productsColly := webColly.Clone()
	productsColly.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 15,
	})
	productsColly.OnHTML("div.list_item", func(el *colly.HTMLElement) {
		el.ForEach("div.item>a", func(_ int, element *colly.HTMLElement) {
			href := element.Attr("href")
			if href != "" {
				productsDetails(webColly, element.Request.AbsoluteURL(href))
			}
		})
	})

	productsColly.OnError(func(resp *colly.Response, err error) {
		log.Println("Category Next: ", resp.Request.URL.String(), "failed with response: ", err)
	})

	productsColly.Visit(url)
}

func productsDetails(webcolly *colly.Collector, url string) {
	products := webcolly.Clone()
	products.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 50,
	})
	products.OnHTML("h1.item-name", func(e *colly.HTMLElement) {
		log.Println(e.Text)
	})

	products.OnError(func(e *colly.Response, err error) {
		log.Println("Products Details Request URL: ", e.Request.URL.String(), "failed with response: ", err)
	})

	products.Visit(url)
}
