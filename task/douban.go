package task

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type Movie struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Director string `json:"director"`
	Author   string `json:"writer"`
	Casts    string `json:"casts"`
	Cate     string `json:"cate"`
	Country  string `json:"country"`
	Language string `json:"language"`
	Release  string `json:"release"`
	Length   string `json:"length"`
	Summary  string `json:"summary"`
	Want     string `json:"want"`
}

const (
	URL     = "https://movie.douban.com/coming"
	REFERER = "https://movie.douban.com/cinema/later/chengdu/"
)

func MovieSchedule() {
	resp, err := getHttpResponser(URL, REFERER)
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan Movie, 1)
	go parseMovieList(resp, &ch)

	done := make(chan Movie, 1)
	go func() {
		for m := range ch {
			if err = parseMoviePages(&m); err != nil {
				log.Fatal(err)
			}
			//done <- m
			break
		}

		close(done)
	}()
}

func getHttpResponser(url, referer string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36")
	request.Header.Add("Referer", referer)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("got status code %d", response.StatusCode))
	}

	if body, err := ioutil.ReadAll(response.Body); err != nil {
		return nil, err
	} else {
		return body, nil
	}
}

func parseMovieList(resp []byte, ch *chan Movie) {
	r, _ := regexp.Compile(
		`<a href="https://movie.douban.com/subject/([0-9]+)/".*?>(.*?)</a>[\s\S]*?` + // Movie id number
			`<td class="">[\s\S]*?([0-9]+)人[\s\S]*?</td>`) // Want watch
	for _, id := range r.FindAllStringSubmatch(string(resp), -1) {
		var m Movie
		m.Id = id[1]
		m.Name = id[2]
		m.Want = id[3]
		*ch <- m
	}

	close(*ch)
}

func parseMoviePages(m *Movie) error {
	page, err := getHttpResponser(
		fmt.Sprintf("https://movie.douban.com/subject/%s/", m.Id), URL)
	if err != nil {
		return err
	}

	r, _ := regexp.Compile(
		`<div id="info">([\s\S]*?)</div>` + // infos
			`[\s\S]*?` +
			`<div class="indent" id="link-report">([\s\S]*?)</div>`) // summary
	items := r.FindStringSubmatch(string(page))
	infoStr := items[1] // Part of informations

	getSummary(&items[2], &m.Summary)
	getInfoA(infoStr, "导演", `<a href.*?>(.*?)</a>`, &m.Director)
	getInfoA(infoStr, "编剧", `<a href.*?>(.*?)</a>`, &m.Author)
	getInfoA(infoStr, "主演", `<a href.*?>(.*?)</a>`, &m.Casts)
	getInfoB(infoStr, "类型:", `<span.*?>(.*?)</span>`, &m.Cate)
	getInfoC(infoStr, "上映日期:", `(\d{4}-\d{2}-\d{2})\(中国大陆\)`, &m.Release)
	getInfoC(infoStr, "片长:", `>(\d+)分钟`, &m.Length)
	getInfoD(infoStr, "制片国家/地区:", `(.*?)`, &m.Country)
	getInfoD(infoStr, "语言:", `(.*?)`, &m.Language)
	//log.Println(m.Summary)

	return nil
}

func getSummary(str, filed *string) {
	var content string

	r, _ := regexp.Compile(`<span property="v:summary" class="">([\s\S]*?)</span>`)
	if items := r.FindStringSubmatch(*str); len(items) != 0 {
		content = items[1]
	} else {
		r, _ := regexp.Compile(`<span class="all hidden">([\s\S]*?)</span>`)
		content = r.FindStringSubmatch(*str)[1]
	}

	strings.Replace(content, " ", "", -1)
	strings.Replace(content, "<br />", "", -1)
	*filed = content
}

// 导演、编剧、主演
func getInfoA(infoStr, name, re string, filed *string) {
	r, _ := regexp.Compile(fmt.Sprintf(`<span class='pl'>%s</span>([\s\S]*?)<br/>`, name))
	filedStr := r.FindStringSubmatch(infoStr)[1]

	r, _ = regexp.Compile(re)
	for n, i := range r.FindAllStringSubmatch(filedStr, -1) {
		*filed += i[1] + "  "
		if n == 2 {
			break
		}
	}
}

// 类型
func getInfoB(infoStr, name, re string, filed *string) {
	r, _ := regexp.Compile(fmt.Sprintf(`<span class="pl">%s</span>([\s\S]*?)<br/>`, name))
	filedStr := r.FindStringSubmatch(infoStr)[1]

	r, _ = regexp.Compile(re)
	for _, i := range r.FindAllStringSubmatch(filedStr, -1) {
		*filed += i[1] + "  "
	}
}

// 上映日期、片长
func getInfoC(infoStr, name, re string, filed *string) {
	r, _ := regexp.Compile(fmt.Sprintf(`<span class="pl">%s</span>([\s\S]*?)<br/>`, name))
	filedStr := r.FindStringSubmatch(infoStr)[1]

	r, _ = regexp.Compile(re)
	if items := r.FindStringSubmatch(filedStr); len(items) != 0 {
		*filed = items[1]
	}
}

// 制片国家、语言
func getInfoD(infoStr, name, re string, filed *string) {
	r, _ := regexp.Compile(fmt.Sprintf(`<span class="pl">%s</span>([\s\S]*?)<br/>`, name))
	*filed = r.FindStringSubmatch(infoStr)[1]
}
