package fetcher

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hi20160616/exhtml"
	"github.com/hi20160616/gears"
	"github.com/hi20160616/ms-trt/configs"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Article struct {
	Id            string
	Title         string
	Content       string
	WebsiteId     string
	WebsiteDomain string
	WebsiteTitle  string
	UpdateTime    *timestamppb.Timestamp
	U             *url.URL
	raw           []byte
	doc           *html.Node
}

var ErrTimeOverDays error = errors.New("article update time out of range")

func NewArticle() *Article {
	return &Article{
		WebsiteDomain: configs.Data.MS.Domain,
		WebsiteTitle:  configs.Data.MS.Title,
		WebsiteId:     fmt.Sprintf("%x", md5.Sum([]byte(configs.Data.MS.Domain))),
	}
}

// List get all articles from database
func (a *Article) List() ([]*Article, error) {
	return load()
}

// Get read database and return the data by rawurl.
func (a *Article) Get(id string) (*Article, error) {
	as, err := load()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		if a.Id == id {
			return a, nil
		}
	}
	return nil, fmt.Errorf("[%s] no article with id: %s, url: %s",
		configs.Data.MS.Title, id, a.U.String())
}

func (a *Article) Search(keyword ...string) ([]*Article, error) {
	as, err := load()
	if err != nil {
		return nil, err
	}

	as2 := []*Article{}
	for _, a := range as {
		for _, v := range keyword {
			v = strings.ToLower(strings.TrimSpace(v))
			switch {
			case a.Id == v:
				as2 = append(as2, a)
			case a.WebsiteId == v:
				as2 = append(as2, a)
			case strings.Contains(strings.ToLower(a.Title), v):
				as2 = append(as2, a)
			case strings.Contains(strings.ToLower(a.Content), v):
				as2 = append(as2, a)
			case strings.Contains(strings.ToLower(a.WebsiteDomain), v):
				as2 = append(as2, a)
			case strings.Contains(strings.ToLower(a.WebsiteTitle), v):
				as2 = append(as2, a)
			}
		}
	}
	return as2, nil
}

type ByUpdateTime []*Article

func (u ByUpdateTime) Len() int      { return len(u) }
func (u ByUpdateTime) Swap(i, j int) { u[i], u[j] = u[j], u[i] }
func (u ByUpdateTime) Less(i, j int) bool {
	return u[i].UpdateTime.AsTime().Before(u[j].UpdateTime.AsTime())
}

var timeout = func() time.Duration {
	t, err := time.ParseDuration(configs.Data.MS.Timeout)
	if err != nil {
		log.Printf("[%s] timeout init error: %v", configs.Data.MS.Title, err)
		return time.Duration(1 * time.Minute)
	}
	return t
}()

// fetchArticle fetch article by rawurl
func (a *Article) fetchArticle(rawurl string) (*Article, error) {
	var err error
	a.U, err = url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	// Dail
	a.raw, a.doc, err = exhtml.GetRawAndDoc(a.U, 1*time.Minute)
	if err != nil {
		if strings.Contains(err.Error(), "invalid header") {
			a.Title = a.U.Path
			a.UpdateTime = timestamppb.Now()
			a.Content, err = a.fmtContent("")
			if err != nil {
				return nil, err
			}
			return a, nil
		} else {
			return nil, err
		}
	}

	a.Id = fmt.Sprintf("%x", md5.Sum([]byte(rawurl)))

	a.Title, err = a.fetchTitle()
	if err != nil {
		return nil, err
	}

	a.UpdateTime, err = a.fetchUpdateTime()
	if err != nil {
		return nil, err
	}

	// content should be the last step to fetch
	a.Content, err = a.fetchContent()
	if err != nil {
		return nil, err
	}

	a.Content, err = a.fmtContent(a.Content)
	if err != nil {
		return nil, err
	}
	return a, nil

}

func (a *Article) fetchTitle() (string, error) {
	n := exhtml.ElementsByTag(a.doc, "title")
	if n == nil {
		return "", fmt.Errorf("[%s] getTitle error, there is no element <title>",
			configs.Data.MS.Title)
	}
	title := n[0].FirstChild.Data
	rp := strings.NewReplacer(" | ???????????????", "", " | ??????", "")
	title = strings.TrimSpace(rp.Replace(title))
	gears.ReplaceIllegalChar(&title)
	return title, nil
}

func (a *Article) fetchUpdateTime() (*timestamppb.Timestamp, error) {
	if a.raw == nil {
		return nil, errors.Errorf("[%s] fetchUpdateTime: raw is nil: %s",
			configs.Data.MS.Title, a.U.String())
	}

	re := regexp.MustCompile(`"datePublished": "(.*?)",`)
	rs := re.FindAllSubmatch(a.raw, -1)
	if rs == nil || len(rs) <= 0 {
		// Just print matched nil to Stderr, instead of return the err
		// so, the news will fetch although datetime err
		fmt.Fprintf(os.Stderr, "[%s] fetchUpdateTime: no datePublished matched: %s",
			configs.Data.MS.Title, a.U.String())
		return timestamppb.Now(), nil
	}
	if len(rs[0]) <= 1 {
		fmt.Fprintf(os.Stderr, "[%s] fetchUpdateTime: no datePublished sub-matched: %s",
			configs.Data.MS.Title, a.U.String())
		return timestamppb.Now(), nil
	}
	t, err := time.Parse(time.RFC3339, string(rs[0][1]))
	if t.Before(time.Now().AddDate(0, 0, -3)) {
		return nil, ErrTimeOverDays
	}
	return timestamppb.New(t), err
}

func shanghai(t time.Time) time.Time {
	loc := time.FixedZone("UTC", 8*60*60)
	return t.In(loc)
}

func (a *Article) fetchContent() (string, error) {
	if a.doc == nil {
		return "", errors.Errorf("[%s] fetchContent: doc is nil: %s", configs.Data.MS.Title, a.U.String())
	}
	body := ""
	// Fetch content nodes
	nodes := exhtml.ElementsByTagAndId(a.doc, "article", "article-body")
	if len(nodes) == 0 {
		nodes = exhtml.ElementsByTagAndClass(a.doc, "div", "article-content-rawhtml")
	}
	if len(nodes) == 0 {
		nodes = exhtml.ElementsByTagAndClass(a.doc, "div", "article-content-container")
	}
	if len(nodes) == 0 {
		return "", errors.Errorf("[%s] no content extract from %s", configs.Data.MS.Title, a.U.String())
	}
	plist := exhtml.ElementsByTag(nodes[0], "p")
	for _, v := range plist {
		if v.FirstChild == nil {
			continue
		} else if v.FirstChild.FirstChild != nil &&
			v.FirstChild.Data == "strong" {
			a := exhtml.ElementsByTag(v, "span")
			for _, aa := range a {
				body += aa.FirstChild.Data
			}
			body += "  \n"
		} else {
			body += v.FirstChild.Data + "  \n"
		}
	}
	body = strings.ReplaceAll(body, "span  \n", "")

	return body, nil
}

func (a *Article) fmtContent(body string) (string, error) {
	var err error
	title := "# " + a.Title + "\n\n"
	lastupdate := shanghai(a.UpdateTime.AsTime()).Format(time.RFC3339)
	webTitle := fmt.Sprintf(" @ [%s](/list/?v=%[1]s): [%[2]s](http://%[2]s)", a.WebsiteTitle, a.WebsiteDomain)
	u, err := url.QueryUnescape(a.U.String())
	if err != nil {
		u = a.U.String() + "\n\nunescape url error:\n" + err.Error()
	}

	body = title +
		"LastUpdate: " + lastupdate +
		webTitle + "\n\n" +
		"---\n" +
		body + "\n\n" +
		"????????????" + fmt.Sprintf("[%s](%[1]s)", u)
	return body, nil
}
