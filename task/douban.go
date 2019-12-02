package task

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mail2ics/clean"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
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
	Url      string `json:"url"`
}

const (
	URL     = "https://movie.douban.com/coming"
	REFERER = "https://movie.douban.com/cinema/later/chengdu/"
)

func MovieSchedule(mc *chan clean.Message) error {
	log.Println("Checking web for movie release info")

	resp, err := getHttpResponser(URL, REFERER)
	if err != nil {
		return err
	}

	ch := make(chan Movie, 1)
	go parseMovieList(resp, &ch)

	done := make(chan Movie, 1)
	go func() {
		for m := range ch {
			if err = parseMoviePages(&m); err != nil {
				log.Fatal(err)
			}
			if strings.Count(m.Release, "-") != 2 {
				continue
			}
			done <- m

			time.Sleep(time.Second * time.Duration(rand.Intn(3)+2))
		}

		close(done)
	}()

	if err = sendToMessage(&done, mc); err != nil {
		return err
	}

	log.Println("Movie release check finished")

	return nil
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
		`<a href="https://movie.douban.com/subject/([0-9]+)/".*?>(.*?)</a>[\s\S]*?` + // Movie's id number and name
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
	m.Url = fmt.Sprintf("https://movie.douban.com/subject/%s/", m.Id)
	page, err := getHttpResponser(m.Url, URL)
	if err != nil {
		return err
	}

	r, _ := regexp.Compile(
		`<div id="info">([\s\S]*?)</div>` + // infos
			`[\s\S]*?` +
			`<div class="indent" id="link-report">([\s\S]*?)</div>`) // summary
	items := r.FindStringSubmatch(string(page))
	infoStr := items[1] // Part of informations

	if err = getSummary(&items[2], &m.Summary); err != nil {
		log.Println(fmt.Sprintf("warning: %s has no info of 'summary'", m.Name))
	}
	if err = getInfoA(infoStr, "导演", `<a href.*?>(.*?)</a>`, &m.Director); err != nil {
		log.Println(fmt.Sprintf("warning: %s has no info of 'director'", m.Name))
	}
	if err = getInfoA(infoStr, "编剧", `<a href.*?>(.*?)</a>`, &m.Author); err != nil {
		log.Println(fmt.Sprintf("warning: %s has no info of 'author'", m.Name))
	}
	if err = getInfoA(infoStr, "主演", `<a href.*?>(.*?)</a>`, &m.Casts); err != nil {
		log.Println(fmt.Sprintf("warning: %s has no info of 'casts'", m.Name))
	}
	if err = getInfoB(infoStr, "类型:", `<span.*?>(.*?)</span>`, &m.Cate); err != nil {
		log.Println(fmt.Sprintf("warning: %s has no info of 'category'", m.Name))
	}
	if err = getInfoC(infoStr, "上映日期:", `content="(.*?)\(中国大陆\)"`, &m.Release); err != nil {
		log.Println(fmt.Sprintf("warning: %s has no info of 'release date'", m.Name))
	}
	if err = getInfoC(infoStr, "片长:", `>(\d+)分钟`, &m.Length); err != nil {
		log.Println(fmt.Sprintf("warning: %s has no info of 'length'", m.Name))
	}
	if err = getInfoD(infoStr, "制片国家/地区:", &m.Country); err != nil {
		log.Println(fmt.Sprintf("warning: %s has no info of 'country/region'", m.Name))
	}
	if err = getInfoD(infoStr, "语言:", &m.Language); err != nil {
		log.Println(fmt.Sprintf("warning: %s has no info of 'language'", m.Name))
	}

	return nil
}

func getSummary(str, filed *string) error {
	r1, _ := regexp.Compile(`<span property="v:summary" class="">([\s\S]*?)</span>`)
	result1 := r1.FindStringSubmatch(*str)
	r2, _ := regexp.Compile(`<span class="all hidden">([\s\S]*?)</span>`)
	result2 := r2.FindStringSubmatch(*str)

	switch {
	case len(result1) != 0:
		*filed = result1[1]
	case len(result2) != 0:
		*filed = result2[1]
	default:
		return errors.New(fmt.Sprintf("can't found info of summary"))
	}

	removeString(filed, " ")
	removeString(filed, "<br/>")
	removeString(filed, "\n")

	return nil
}

func removeString(str *string, remove string) {
	b := []byte(*str)
	re := regexp.MustCompile(remove)
	*str = string(re.ReplaceAll(bytes.TrimSpace(b), []byte("")))
}

// 导演、编剧、主演
func getInfoA(infoStr, name, re string, filed *string) error {
	var filedStr string
	r, _ := regexp.Compile(fmt.Sprintf(`<span class='pl'>%s</span>([\s\S]*?)<br/>`, name))
	if result := r.FindStringSubmatch(infoStr); len(result) != 0 {
		filedStr = result[1]
	} else {
		return errors.New(fmt.Sprintf("can't found info of %s", name))
	}

	r, _ = regexp.Compile(re)
	for n, i := range r.FindAllStringSubmatch(filedStr, -1) {
		*filed += i[1] + "  "
		if n == 2 {
			break
		}
	}

	return nil
}

// 类型
func getInfoB(infoStr, name, re string, filed *string) error {
	var filedStr string
	r, _ := regexp.Compile(fmt.Sprintf(`<span class="pl">%s</span>([\s\S]*?)<br/>`, name))
	if result := r.FindStringSubmatch(infoStr); len(result) != 0 {
		filedStr = result[1]
	} else {
		return errors.New(fmt.Sprintf("can't found info of %s", name))
	}

	r, _ = regexp.Compile(re)
	for _, i := range r.FindAllStringSubmatch(filedStr, -1) {
		*filed += i[1] + "  "
	}

	return nil
}

// 上映日期、片长
func getInfoC(infoStr, name, re string, filed *string) error {
	var filedStr string
	r, _ := regexp.Compile(fmt.Sprintf(`<span class="pl">%s</span>([\s\S]*?)<br/>`, name))
	if result := r.FindStringSubmatch(infoStr); len(result) != 0 {
		filedStr = result[1]
	} else {
		return errors.New(fmt.Sprintf("can't found info of %s", name))
	}

	r, _ = regexp.Compile(re)
	if items := r.FindStringSubmatch(filedStr); len(items) != 0 {
		*filed = items[1]
	}

	return nil
}

// 制片国家、语言
func getInfoD(infoStr, name string, filed *string) error {
	r, _ := regexp.Compile(fmt.Sprintf(`<span class="pl">%s</span>([\s\S]*?)<br/>`, name))
	if result := r.FindStringSubmatch(infoStr); len(result) != 0 {
		*filed = result[1]
	} else {
		return errors.New(fmt.Sprintf("can't found info of %s", name))
	}

	return nil
}

func sendToMessage(done *chan Movie, mc *chan clean.Message) error {
	var msg clean.Message
	events := make([]clean.Event, 100)

	msg.From = "brucegxs@gmail.com"
	msg.Subject = fmt.Sprintf("%s 电影上映时间表更新", time.Now().Format("2006-01-02"))
	msg.Cal = "电影时间表"
	msg.Filename = "movie.ics"
	msg.Events = events

	index := 0
	for m := range *done {
		if err := eventAssignment(&events[index], &m); err != nil {
			return err
		}
		index++
	}

	*mc <- msg

	return nil
}

func eventAssignment(event *clean.Event, m *Movie) error {
	// Summary
	event.Summary = fmt.Sprintf("%s, %s人想看", m.Name, m.Want)
	// start time and end time
	if err := getStartAndEndDate(m, event); err != nil {
		return err
	}
	// Uid
	event.Uid = fmt.Sprintf("mvrelease"+"%s", m.Id)
	// Detail
	event.Detail = fmt.Sprintf(
		"导演: %s\\n"+
			"编剧: %s\\n"+
			"主演: %s\\n"+
			"类型: %s\\n"+
			"国家/地区: %s\\n"+
			"语言: %s\\n"+
			"片长: %s\\n"+
			"剧情简介: %s\\n"+
			"网址: %s", m.Director, m.Author, m.Casts, m.Cate, m.Country, m.Language, m.Length, m.Summary, m.Url)

	return nil
}

func getStartAndEndDate(m *Movie, event *clean.Event) error {
	if st, err := clean.ParseTime(m.Release, "2006-01-02", 0, "-"); err != nil {
		return err
	} else {
		if et, err := clean.ParseTime(st, clean.ICS_DT, 24, "+"); err != nil {
			return err
		} else {
			event.EndDT = fmt.Sprintf(";VALUE=DATE:%s", strings.Split(et, "T")[0])
		}

		event.StartDT = fmt.Sprintf(";VALUE=DATE:%s", strings.Split(st, "T")[0])
	}

	return nil
}
