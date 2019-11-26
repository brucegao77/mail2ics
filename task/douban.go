package task

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

type list struct {
	id   int
	want int
}

type Movie struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Director string `json:"director"`
	Writer   string `json:"writer"`
	Casts    string `json:"casts"`
	Cate     string `json:"cate"`
	Country  string `json:"country"`
	Language string `json:"language"`
	Release  string `json:"release"`
	Length   string `json:"length"`
	Alias    string `json:"alias"`
	Summary  string `json:"summary"`
	Want     string `json:"want"`
}

const (
	URL     = "https://movie.douban.com/coming"
	REFERER = "https://movie.douban.com/cinema/later/chengdu/"
)

func MovieSchedule() {
	ch := make(chan Movie, 1)

	resp, err := getHttpResponser(URL, REFERER)
	if err != nil {
		log.Fatal(err)
	}

	go parseMovieList(resp, &ch)

	done := make(chan Movie, 1)
	go func() {
		for m := range ch {
			if err = parseMoviePages(&m, &done); err != nil {
				log.Fatal(err)
			}
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
		`<a href="https://movie.douban.com/subject/([0-9]+)/"[\s\S]*?<td class="">[\s\S]*?([0-9]+)人[\s\S]*?</td>`)
	for _, id := range r.FindAllStringSubmatch(string(resp), -1) {
		var m Movie
		m.Id = id[1]
		m.Want = id[2]
		*ch <- m
	}

	close(*ch)
}

func parseMoviePages(m *Movie, done *chan Movie) error {
	page, err := getHttpResponser(
		fmt.Sprintf("https://movie.douban.com/subject/%s/", m.Id), URL)
	if err != nil {
		return err
	}

	return nil
}
