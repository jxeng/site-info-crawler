package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jxeng/site-info-crawler/tool"
	"github.com/jxeng/site-info-crawler/types"

	"github.com/chromedp/chromedp"
)

var ctx context.Context

func main() {
	var items []types.Item

	tool.ReadJsonFile("./raw.json", &items)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("ignore-certificate-errors", true),
	)

	var cancel context.CancelFunc

	ctx, _ = chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	for i := range items {
		err := crawl(&items[i])
		log.Println(items[i].Url, err)
	}

	tool.WriteJsonFile("./filled.json", &items)
}

func crawl(item *types.Item) error {
	var urlstr = item.Url

	var title = ""
	var desc = ""
	var iconPath = ""
	var iconExists = false

	iconFile := "./icons/" + item.Id
	if _, err := os.Stat(iconFile + ".png"); err == nil {
		iconExists = true
	}
	if _, err := os.Stat(iconFile + ".svg"); err == nil {
		iconExists = true
	}

	if item.Description != "" && iconExists {
		return nil
	}

	// use index.html to fetch site info
	var iconJs = `
const iconElement = document.querySelector("link[rel~=icon]");
const iconPath = (iconElement && iconElement.href) || '';
iconPath
`
	var descJs = `
let descElement = document.querySelector('meta[name="description"]');
if (!descElement) {
	descElement = document.querySelector('meta[property="og:description"]');
}
const desc = (descElement && descElement.content) || "";
desc
`

	resp, err := tool.Request(urlstr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		if buf, err := ioutil.ReadAll(resp.Body); err == nil {
			// title info
			r, _ := regexp.Compile(`<title>(.+?)<\/title>`)
			res := r.FindSubmatch(buf)
			if len(res) > 1 {
				title = string(res[1])
			}

			// description
			r, _ = regexp.Compile(`<meta.+?(?:name|property)="(?:og:)?description" content="(.+?)"`)
			res = r.FindSubmatch(buf)
			if len(res) > 1 {
				desc = string(res[1])
			}

			// favicon
			r, _ = regexp.Compile(`<link.+?rel=".{0,20}?icon.{0,20}?".+?href="(.+?)"`)
			res = r.FindSubmatch(buf)
			if len(res) > 1 {
				iconPath = string(res[1])
			}
			if iconPath == "" {
				r, _ = regexp.Compile(`<link.+?href="(.+?)".+?rel=".{0,20}?icon.{0,20}?"`)
				res = r.FindSubmatch(buf)
				if len(res) > 1 {
					iconPath = string(res[1])
				}
			}
		} else {
			log.Println(err)
		}
	} else {
		log.Println(resp.Status)
	}

	// use chromedp to fetch site info
	if title == "" || desc == "" || iconPath == "" {
		if err := chromedp.Run(ctx,
			chromedp.Navigate(urlstr),
			chromedp.Sleep(5*time.Second),
		); err != nil {
			log.Println("chromedp error:", err)
		}

		if title == "" {
			chromedp.Run(ctx, chromedp.Title(&title))
		}
		if iconPath == "" {
			chromedp.Run(ctx, chromedp.Evaluate(iconJs, &iconPath))
		}
		if desc == "" {
			chromedp.Run(ctx, chromedp.Evaluate(descJs, &desc))
		}
	}

	tmp := strings.Split(string(title), " - ")
	title = strings.Trim(tmp[0], " ")
	if len(tmp) > 1 && desc == "" {
		desc = strings.Trim(tmp[1], " ")
	}
	item.Title = title
	item.Description = desc

	if iconExists {
		return nil
	}

	u, err := url.Parse(urlstr)
	if err != nil {
		return err
	}

	// favicon download url
	var dlUrl string
	if iconPath == "" {
		dlUrl = fmt.Sprintf("%s://%s/favicon.ico", u.Scheme, u.Host)
	} else if strings.HasPrefix(iconPath, "data:image") {
		iconFile += ".png"
		item.Favicon = iconFile
		if err = tool.SaveIcon(iconPath, iconFile); err != nil {
			return err
		}
		return nil
	} else if strings.HasPrefix(iconPath, "http") {
		dlUrl = iconPath
	} else if strings.HasPrefix(iconPath, "//") {
		dlUrl = "https:" + iconPath
	} else if strings.HasPrefix(iconPath, "/") {
		dlUrl = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, iconPath)
	} else if strings.HasPrefix(iconPath, "../") {
		fmt.Println(u.Host, u.Path)
		dlUrl = fmt.Sprintf("%s://%s%s/%s", u.Scheme, u.Host, u.Path, iconPath)
	} else {
		path := ""
		index := strings.LastIndex(u.Path, "/")
		if index > -1 {
			path = path[0:index]
		}
		dlUrl = fmt.Sprintf("%s://%s%s/%s", u.Scheme, u.Host, path, strings.TrimLeft(strings.TrimLeft(iconPath, "."), "/"))
	}

	if strings.HasSuffix(dlUrl, ".svg") {
		iconFile += ".svg"
	} else {
		iconFile += ".png"
	}

	item.Favicon = iconFile
	if err = tool.Download(dlUrl, iconFile); err != nil {
		return err
	}

	return nil
}
