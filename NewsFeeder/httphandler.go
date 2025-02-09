package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/xml"
	"sync"
	"bytes"
	"strings"
	"html/template"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"os"
)



var news map[string] NewsInfo = make(map[string] NewsInfo)
var newsfeeder *Newsfeeder = nil
var urls_list map[string] string = make(map[string]string)

func index_handler(w http.ResponseWriter, r *http.Request) {
	html := "<h1> News aggregator for stocks market magazines </h1>"
	fmt.Fprintf(w, html)
}

func newsaggregator(w http.ResponseWriter, r *http.Request){
	data := struct {
		Title string
		NewsList map[string] NewsInfo
		Urls map[string] string
	}{
		Title : "News aggregator", 
		NewsList : news,
		Urls: urls_list,
	}

	// funcmap := template.FuncMap{
	// 	"add": func(a, b int) int {
	// 		return a+b
	// 	},
	// }
	//fmt.Println(news)
	//funcmap := template
	html, _ := template.ParseFiles("newsfeeder.html")
	//html = html.Funcs(funcmap)
	fmt.Println(html.Execute(w, data))
}

//xml data structure to parse and get the xml feed

type Sitemapindex struct{
	Locations string `xml:"channel>link"`
	News []NewsData `xml:"channel>item"`
}

type NewsData struct{
	Title string `xml:"title"`
	Description string `xml:"description"`
	Link string `xml:"link"`
	Category string `xml:"category"`
	PublishDate string `xml:"pubDate"`
}

//data for news filling

type NewsInfo struct{
	Title string 
	Description string
	Category string
	Source string
	Link string
	PublishDate string
	Important string
	Tracked string
	Logic string
}

func (n NewsInfo) String() string {
	fmt.Printf("Title - %s\n", n.Title)
	fmt.Printf("Description - %s\n", n.Description)
	fmt.Printf("Link- %s\n", n.Link)
	fmt.Printf("Category - %s\n", n.Category)
	fmt.Printf("Published date - %s\n", n.PublishDate)
	fmt.Printf("Source of this news -%s\n", n.Source)
	return fmt.Sprintf("----------------------------------------------------------------\n")
}

func IsImportantNews(title, description string) (bool, string) {
	//get the keywords from db to see any important keywords present in title or descriptions
	for _, k := range newsfeeder.Keywords{
		if strings.Contains(title, k) || strings.Contains(description, k){
			return true, k
		}
	}
	
	return false, "False"
}


func get_news(wg *sync.WaitGroup, ch chan NewsInfo, url string){
	defer wg.Done()
	
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil{
		fmt.Println("Error occoured while creating new request for url ", url)
		var message string
		fmt.Sprintf(message, err)
		urls_list[url] = message
		return
	}

	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil{
		fmt.Println("Error occured while sending client get request")
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(data))
	data = bytes.Trim(data, `" `)
	var s Sitemapindex
	xml.Unmarshal(data, &s)
	//fmt.Println(url, "info 4")
	//fmt.Println(s)
	resp.Body.Close()

	if len(s.News) == 0{
		var message string
		message = fmt.Sprintf("no message retrieved from this url - %s", url)
		urls_list[url] = message
		fmt.Println("No news info got from this url  - ", url)
		return
	}

	for _, n := range s.News{
		

		newsinfo := NewsInfo{n.Title, n.Description, n.Category, url, n.Link, n.PublishDate, "False", "False", " "}
		for _, c := range newsfeeder.Companies{
			if strings.Contains(n.Title, c.Name) || strings.Contains(n.Description, c.Name){
				newsinfo.Tracked = fmt.Sprintf("** - %s", c.Category)
				newsinfo.Logic = fmt.Sprintf("** - %s", c.Logic)								
				break
			}
		}

		var important string
		var ok bool
		
		if ok, important = IsImportantNews(n.Title, n.Description); ok == true {
			newsinfo.Important = fmt.Sprintf("** - %s", important)
		} else{
			newsinfo.Important = important
		}		

		ch <- newsinfo
	}
	urls_list[url] = "successfully retrieved all the news"
}

type Config struct{
	 Newsfeed Newsfeeder `json:"newsfeeder"`
}

type Newsfeeder struct{
	Urls [] string `json:"urls"`
	Companies [] Company `json:"companies"`
	Keywords [] string `json:"keywords"`
}

type Company struct{
	Name string `json:"name"`
	Category string `json:"category"`
	Logic string `json:"logic"`
}

func ParseJSONConfig() (*Newsfeeder, error){
	//fmt.Println("ParseJSONconfig function 1")
	jsonfile, err := os.Open("newsfeeder.json")
	if err != nil{
		fmt.Println("error reading file")
		error := fmt.Sprintf("Failed to open newsfeeder JSON file %W", err)
		fmt.Println(error, err)
		return nil, errors.New(error)
	}
	//fmt.Println("ParseJSONconfig function 1")

	defer jsonfile.Close()

	bytevalue, err:= ioutil.ReadAll(jsonfile)
	if err != nil{
		fmt.Println("reading JSON file failed with error ", err)
		return nil, errors.New("reading JSON file failed with error ")
	}

	//fmt.Println(bytevalue)
	var config Config
	error_mar := json.Unmarshal(bytevalue, &config)
	if error_mar !=nil{
		fmt.Println("unmarshal error", error_mar.Error())
	}
	//fmt.Println(config)
	importantnews := config.Newsfeed
	return &importantnews, nil
}


func main() {

	

	//Get all the news feeds and store in map key title and value NewsInfo struct
	var wg sync.WaitGroup

	config, err := ParseJSONConfig()
	if err != nil{
		fmt.Println(err)
	}
	newsfeeder = config

	mmap := make(map[int] string)

	fmt.Println("getting stock news from urls listed below")
	for index, url := range newsfeeder.Urls{
		fmt.Println(url)
		mmap[index] = url
	}

	ch := make(chan NewsInfo, 50)

	for _, url := range mmap{
		wg.Add(1)
		go get_news(&wg, ch, url)
	}

	go func(){
		wg.Wait()
		close(ch)
		fmt.Println("channel closed")
	}()

    //fmt.Println("info 6")
	for n := range ch{
		//fmt.Printf("%T\n", n)
		//store in db
		id := sha256.New()
		id.Write([]byte(n.Title))
		//fmt.Println("getting news", n)
		//fmt.Printf("getting data %T, %s", news, string(id.Sum(nil)))
		news[string(id.Sum(nil))] = n
		//fmt.Println(n)
	}

	//fmt.Println(news)
	//fmt.Println("info 7")

	http.HandleFunc("/", index_handler)
	http.HandleFunc("/stocks", newsaggregator)
	http.ListenAndServe(":8000", nil)
	
}