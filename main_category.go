package main

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/xuri/excelize/v2"
)

func main() {
	browserPath := "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
	u := launcher.New().Bin(browserPath).MustLaunch()
	broswer := rod.New().ControlURL(u).MustConnect()
	page := broswer.MustPage("https://electricbikereview.com/category/")
	defer broswer.MustClose()
	page.MustWaitDOMStable()
	categories := page.MustElementsX("/html/body/div[1]/section/div[2]/div/div[1]/div[2]/div/div/div/a/@href")
	categories_links := make([]string, len(categories))
	base_url := "https://electricbikereview.com"
	for i, category := range categories {
		categories_links[i] = base_url + category.MustText()
	} //获取所有分类的链接
	//var lock sync.Mutex
	for i, category_link := range categories_links {
		if i < 1 {
			continue
		}
		name := category_link[40 : len(category_link)-1]
		fmt.Println("gogogo!!!!!!!!!!!!!!!!!!!!!!!!")
		f := excelize.NewFile()
		f.Path = name + ".xlsx"
		_, err := f.NewSheet(name)
		if err != nil {
			panic(err)
		}
		if len(name) > 30 {
			name = name[:30]
		}
		Go_details(name, f, broswer, category_link)
		f.Save()
		f.Close()
		fmt.Println("a category done!!!!!!!!!!!!!!!!!!!!!!!!")
	}
}
