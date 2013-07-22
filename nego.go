package main

import (
	"code.google.com/p/go.net/html"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
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
	time              time.Time
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
		p.tmpLink.Text = n.Data
		p.Links = append(p.Links, p.tmpLink)
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

	p.time = time.Now()
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

	//$test=$site["name"]."\n";#"testname\n";
	//$test.="name:".$site["name"]."\n";
	//$test.="cols:".$site["colcount"]."\n";
	//$test.="headln:".trim($site->headline)."\n";
	//$test.="logo:".trim($site["logo"])."\n";
	//$test.="url:".trim($site->siteurl)."\n";

	// $test.=$links;
	postUrl := "http://www.newsempire.net/up/shownews.php"
	// Submit form
	resp, err := http.PostForm(postUrl, values)
	if err != nil {
		//log.Fatal(err)
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
		fmt.Printf("Link <%s> href <%s> Date <%s>\n", v.Text, v.Href, RenderTime(p.time))
	}

}
