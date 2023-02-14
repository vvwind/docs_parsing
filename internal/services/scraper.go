package services

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/spf13/viper"
	"strings"
)

type Scraper struct {
	Data []string
}

func (s *Scraper) Start() error {
	var err error
	data := []string{}
	c := colly.NewCollector(
		colly.AllowedDomains(viper.GetString("allowedDomain")),
	)
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("Visiting %s\n", r.URL)
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Printf("%v", e)
		err = e
	})
	flag := false
	c.OnHTML("tr td", func(e *colly.HTMLElement) {
		result := e.Text
		tags := ""
		e.ForEach("ul li", func(i int, el *colly.HTMLElement) {
			flag = true
			tags += "* " + el.Text + "\n"
			result = strings.Replace(result, el.Text, "", 1)
		})
		if !flag {
			data = append(data, result+"\n")
		} else {
			data = append(data, result+"\n"+tags)
		}

	})
	errorVisiting := c.Visit(viper.GetString("scrapURL"))
	if errorVisiting != nil {
		err = errorVisiting
	}
	s.Data = data
	return err
}
