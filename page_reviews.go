package main

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/xuri/excelize/v2"
	"strconv"
	"strings"
	"sync"
)

func Go_details(sheet string, f *excelize.File, broswer *rod.Browser, url string) {
	if strings.Contains(url, "other-road") || strings.Contains(url, "other-trike") {
		return //这两个分类下的车辆是空，不爬取
	}
	var lock sync.Mutex
	page := broswer.MustPage(url)
	fmt.Println("=====================================")
	fmt.Println("page_index:", url)
	page.MustWaitDOMStable()
	defer page.MustClose()
	page.WaitElementsMoreThan("body > div.page-wrapper > section > div.main-content > div > div.left-side-content > div.product-card-wrapper > div > div:nth-child(4) > div > div.owl-stage-outer > div > div:nth-child(4) > a", 1)
	details := page.MustElementsX("/html/body/div[1]/section/div[2]/div/div[1]/div[2]/div/div[3]/div/div[1]/div/div[4]/a/@href")
	details_links := make([]string, len(details))
	for i, detail := range details {
		details_links[i] = "https://electricbikereview.com" + detail.MustText()
	} //获取每个链接
	//爬取该品牌所有的页数
	for page_index := page.MustElementX("/html/body/div[1]/section/div[2]/div/div[1]/ul/li[last()]/a/@href").MustText(); strings.Contains(page_index, "page"); {
		fmt.Println("=====================================")
		fmt.Println("page_index:", page_index)
		page = broswer.MustPage(url + page_index)
		page.WaitElementsMoreThan("body > div.page-wrapper > section > div.main-content > div > div.left-side-content > div.product-card-wrapper > div > div:nth-child(4) > div > div.owl-stage-outer > div > div:nth-child(4) > a", 1)
		page.MustWaitDOMStable()
		details := page.MustElementsX("/html/body/div[1]/section/div[2]/div/div[1]/div[2]/div/div[3]/div/div[1]/div/div[4]/a/@href")
		for _, detail := range details {
			details_links = append(details_links, "https://electricbikereview.com"+detail.MustText())
		}
		page_index = page.MustElementX("/html/body/div[1]/section/div[2]/div/div[1]/ul/li[last()]/a/@href").MustText()
		fmt.Println("details_links:", len(details_links))
	}
	//开始爬取每个链接的内容
	//通过工作池并发爬取
	works := make(chan int, 5)
	var wg sync.WaitGroup
	for i, detail_link := range details_links {
		works <- 1
		wg.Add(1)
		fmt.Println("detail_link:", detail_link, i)
		sv := 0
		if i%20 == 0 {
			sv = 1
		}
		go Details_get(sv, &lock, strconv.Itoa(i+2), sheet, f, broswer, detail_link, works, &wg)
	}
	wg.Wait()
	close(works)
}
