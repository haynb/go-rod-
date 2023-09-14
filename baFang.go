package main

import (
	"database/sql"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

func main() {
	browserPath := "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
	u := launcher.New().Bin(browserPath).MustLaunch()
	broswer := rod.New().ControlURL(u).MustConnect()
	defer broswer.MustClose()
	baseUrl := "https://www.bafang-e.com/"
	page := broswer.MustPage("https://www.bafang-e.com/cn/oem-area/components/motor/m-series")
	db, err := sql.Open("mysql", "root:heanyang@tcp(10.199.1.41:8848)/motor?charset=utf8mb4&parseTime=True")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	defer db.Close()
	page.MustWaitDOMStable()
	//tableName := "中置"
	var motorList []string
	motorLinks := page.MustElementsX("//*[@id=\"productUI\"]/div/@onclick")
	for _, motorLink := range motorLinks {
		motorList = append(motorList, baseUrl+strings.Split(motorLink.MustText(), "'")[1])
	}
	fmt.Println("motorList:    ", len(motorList))
	var left string
	leftList := make(map[string]bool)
	for _, motor := range motorList {
		page = broswer.MustPage(motor)
		page.MustWaitLoad()
		lists := page.MustElementsX("/html/body/main/div/div[1]/div[2]/div[2]/div[1]/div/p/i")
		for _, list := range lists {
			left = list.MustText()
			fmt.Println(left)
			if _, ok := leftList[left]; !ok {
				leftList[left] = true
			} else {
				continue
			}
		}
	}
	for key := range leftList {
		fmt.Println(key)
	}
}
