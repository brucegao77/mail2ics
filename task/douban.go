package task

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

type list struct {
	id   int
	want int
}

type Movie struct {
	Id       int    `json:"id"`
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

func GetHttpResponser(url, referer string) ([]byte, error) {
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

func ParseComingList(resp []byte) {
	fmt.Println(string(resp))
	r, _ := regexp.Compile(
		`<a href="https://movie.douban.com/subject/([0-9]+)/"[\s\S]*?<td class="">[\s\S]*?([0-9]+)人[\s\S]*?</td>`)
	for _, id := range r.FindAllStringSubmatch(string(resp), -1) {
		fmt.Println(id)
	}
}
