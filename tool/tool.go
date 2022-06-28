package tool

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jxeng/site-info-crawler/types"
)

func ReadJsonFile(path string, items *[]types.Item) {
	jsonStr, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
		return
	}

	err = json.Unmarshal(jsonStr, &items)
	if err != nil {
		log.Fatalln(err)
	}
}

func WriteJsonFile(path string, obj interface{}) {
	str, err := json.Marshal(obj)
	if err != nil {
		log.Fatalln(err)
	}
	err = ioutil.WriteFile(path, str, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func SaveIcon(data, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	i := strings.Index(data, ",")
	if i < 0 {
		log.Fatal("no comma")
	}
	dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data[i+1:]))
	_, err = io.Copy(f, dec)
	if err != nil {
		return err
	}
	return nil
}

func Request(url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.64 Safari/537.36 Edg/101.0.1210.53")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	return resp, nil
}

func Download(url, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	resp, err := Request(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
