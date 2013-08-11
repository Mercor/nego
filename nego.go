package main

import (
	"bytes"
	"code.google.com/p/go.net/html"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type Link struct {
	Href string
	Text string
}

type Page struct {
	SiteURL           string
	baseURL           *url.URL
	LinkRegex         string
	linkRegexCompiled *regexp.Regexp
	foundLink         bool
	Links             []Link
	tmpLink           Link
	Doc               *html.Node
	Time              time.Time
}

func (p *Page) getLinks() {
	fmt.Println("Get Links ", p.Doc)
	p.getLink(p.Doc)
}

func (p *Page) getLink(n *html.Node) {

	if n == nil {
		return
	}
	if n.Type == html.TextNode && p.foundLink == true {
		fmt.Println("found Text Token ", n.Data)
		p.tmpLink.Text = strings.TrimSpace(n.Data)
		if len(p.tmpLink.Text) > 3 {
			p.Links = append(p.Links, p.tmpLink)
		}
		p.foundLink = false
	}
	if n.Type == html.ElementNode && n.Data == "a" {

		for _, a := range n.Attr {
			if a.Key == "href" {
				fmt.Printf("found link  %s  ", a.Val)
				if p.linkRegexCompiled.MatchString(a.Val) {
					p.foundLink = true
					u, err := url.Parse(a.Val)
					if err != nil {
						fmt.Println("error parsing url ", err)
					}
					if !u.IsAbs() {
						p.tmpLink.Href = p.baseURL.ResolveReference(u).String()
					} else {
						p.tmpLink.Href = a.Val

					}
					fmt.Printf("match !\n")
					break
				}
				fmt.Printf("no match...\n")

			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		//fmt.Println(c)
		p.getLink(c)
	}
}

func (p *Page) LoadSite() {

	p.Time = time.Now()
	p.baseURL, _ = url.Parse(p.SiteURL)
	client := &http.Client{}
	client.CheckRedirect =
		func(req *http.Request, via []*http.Request) error {
			fmt.Fprintf(os.Stderr, "Redirect: %v\n", req.URL)
			return nil
		}

	page, err := client.Get(p.SiteURL)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	p.Doc, err = html.Parse(page.Body)

	if err != nil {
		fmt.Print(err)
	}
	fmt.Println("LoadSite ", p.Doc)
	//	page.Body.Close()

	p.linkRegexCompiled = regexp.MustCompile(p.LinkRegex)

}

func (p *Page) PostForm() {

	values := make(url.Values)

	values.Set("submitted", "true")
	//	values.Set("news", $data)

	var buffer bytes.Buffer

	fmt.Println(buffer.String())
	buffer.WriteString("test\n")
	buffer.WriteString("name:test\n")
	buffer.WriteString("cols:2\n")
	buffer.WriteString("headln:test\n")
	buffer.WriteString("logo:test.jpg\n")
	buffer.WriteString("url:test.com\n")
	for _, i := range p.Links {
		buffer.WriteString(i.Text)
		buffer.WriteString("|")
		buffer.WriteString(i.Href)
		buffer.WriteString("|")
		buffer.WriteString(RenderTime(p.Time))
		buffer.WriteString("\n")

	}
	values.Set("news", buffer.String())

	proxyUrl, err := url.Parse("http://127.0.0.1:8888")
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}

	postUrl := "http://www.newsempire.net/up/shownews.php"
	// Submit form
	resp, err := myClient.PostForm(postUrl, values)
	if err != nil {
		fmt.Println(err)
	}

	resp.Body.Close()

}

func RenderTime(t time.Time) string {
	return fmt.Sprintf("%d%02d%02d %02d:%02d",
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute())
}

func main() {

	var p Page
	p.SiteURL = "http://entwickler.com/aggregator/categories/1"
	p.LinkRegex = ".*news.*|webmagazin.de"
	p.LoadSite()
	p.getLinks()
	for _, v := range p.Links {
		fmt.Printf("Link <%s> href <%s> Date <%s>\n", v.Text, v.Href, RenderTime(p.Time))
	}
	p.PostForm()
}
