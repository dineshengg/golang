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
// 	"database/sql"
// 	"github.com/uptrace/bun"
// 	"github.com/uptrace/bun/extra/bundebug"
// 	"github.com/uptrace/bun/dialect/pgdialect"
// 	"github.com/uptrace/bun/driver/pgdriver"
)

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

type NewsInfo struct{
	Title string 
	Description string
	Category string
	Source string
	Link string
	PublishDate string
	Important bool
}

var news map[string] NewsInfo = make(map[string] NewsInfo)

func index_handler(w http.ResponseWriter, r *http.Request) {
	html := "<h1> News aggregator for stocks market magazines </h1>"
	fmt.Fprintf(w, html)
}

func newsaggregator(w http.ResponseWriter, r *http.Request){
	data := struct {
		Title string
		NewsList map[string] NewsInfo
	}{
		Title : "News aggregator", 
		NewsList : news,
	}
	fmt.Println(news)
	//funcmap := template
	html, _ := template.ParseFiles("newsfeeder.html")
	fmt.Println(html.Execute(w, data))
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

func IsImportantNews(title, description string) bool {
	//get the keywords from db to see any important keywords present in title or descriptions
	Keywords := [] string {"bank", "government", "order", "bags", "contract", "worth"}
	for _, k := range Keywords{
		if strings.Contains(title, k) || strings.Contains(description, k){
			return true
		}
	}
	return false
}


//New feed url list data structure
// type NewsUrl struct{
// 	bun.BaseModel `bun:"table:users,alias:u"`

// 	ID int64 `bun:",pk,autoincrement"`
// 	URL string
// }

//var dsn string = "postgres://postgre:admin@NewsFeeds/var/run/postgresql/.s.PGSQL.5432"

// func (p *NewsUrl) database_createtable(db *DB) int{
// 	res, err := db.NewCreateTable().Model(&NewsUrl).Exec(ctx)
// 	if err != nil{
// 		fmt.Println("Unable to create table for NewsUrl")
// 		return -1
// 	}
    
//     //captures query errors in stdout
// 	db.AddQueryHook(bundebug.NewQueryHook(
// 	bundebug.WithVerbose(true),
// 	bundebug.FromEnv("BUNDEBUG"),
// ))
// }

// func (p * NewsUrl) query_newsurls(db *DB) [] NewsUrl{
// 	newsurls = [] NewsUrl
// 	err := db.NewSelect().Model(&newsurls).Scan(ctx)
// 	if err != nil{
// 		fmt.Println("Unable to create table for NewsUrl")
// 		return nil
// 	}
// 	return newsurls
// }


func get_news(wg *sync.WaitGroup, ch chan NewsInfo, url string){
	defer wg.Done()
	fmt.Println("info 1")
	//resp, err:= http.Get(url)
	fmt.Println("info 2")
	// if err != nil{
	// 	fmt.Printf("Not able to get RSS feed - %s", err)
	// 	return
	// }

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil{
		fmt.Println("Error occoured while creating new request for url ", url)
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
	fmt.Println("info 4")
	//fmt.Println(s)
	resp.Body.Close()

	if len(s.News) == 0{
		fmt.Println("No news info got from this url  - ", url)
		return
	}

	for _, n := range s.News{
		newsinfo := NewsInfo{n.Title, n.Description, n.Category, url, n.Link, n.PublishDate, false}
		//fmt.Println(newsinfo)
		//newsinfo.Important = IsImportantNews(n.Title, n.Description)
		ch <- newsinfo
	}
	
	fmt.Println("info 5")
	
}



func main() {

	// var newsurl NewsUrl
	// sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	// db := bun.NewDB(sqldb, pgdialect.New())
	// newsurl.database_createtable(&db)
	// newsurls := newsurl.query_newsurls(&db)

	//Get all the news feeds and store in map key title and value NewsInfo struct
	var wg sync.WaitGroup

	

	mmap := make(map[int] string)
	mmap[0] = "https://www.thehindubusinessline.com/news/feeder/default.rss"
	//mmap[1] = "https://www.business-standard.com/rss/markets-106.rss"
	//Economic times
	// mmap[2] = "https://cfo.economictimes.indiatimes.com/rss/economy"
	// mmap[3] = "https://cfo.economictimes.indiatimes.com/rss/corporate-finance"
	// mmap[4] = "https://cfo.economictimes.indiatimes.com/rss/topstories"
	// mmap[5] = "https://cfo.economictimes.indiatimes.com/rss/policy"
	// mmap[6] = "https://cfo.economictimes.indiatimes.com/rss/governance-risk-compliance"
	// mmap[7] = "https://cfo.economictimes.indiatimes.com/rss/lateststories"
	// mmap[8] = "https://www.livemint.com/rss/markets"
	// mmap[9] = "https://www.livemint.com/rss/companies"
	// mmap[10] = "https://www.livemint.com/rss/industry"
	ch := make(chan NewsInfo, 50)

	for _, url := range mmap{
		wg.Add(1)
		go get_news(&wg, ch, url)
	}

	go func(){
		wg.Wait()
		close(ch)
		fmt.Println("closed channel")
	}()

    fmt.Println("info 6")
	for n := range ch{
		//fmt.Printf("%T\n", n)
		//store in db
		id := sha256.New()
		id.Write([]byte(n.Title))
		//fmt.Println("getting news", n)
		//fmt.Printf("getting data %T, %s", news, string(id.Sum(nil)))
		news[string(id.Sum(nil))] = n
	}

	fmt.Println(news)
	fmt.Println("info 7")

	http.HandleFunc("/", index_handler)
	http.HandleFunc("/stocks", newsaggregator)
	http.ListenAndServe(":8000", nil)
	
}