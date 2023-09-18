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
	page := broswer.MustPage("https://www.bafang-e.com/en/oem-area/components/motor/hf-series")
	page.MustWaitDOMStable()
	db, err := sql.Open("mysql", "root:heanyang@tcp(10.199.1.41:8848)/爬虫?charset=utf8mb4&parseTime=True")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	defer db.Close()
	page.MustWaitDOMStable()
	var (
		motorList []string
		img       []string
	)
	motorLinks := page.MustElementsX("//*[@id=\"productUI\"]/div/@onclick")
	for _, motorLink := range motorLinks {
		motorList = append(motorList, baseUrl+strings.Split(motorLink.MustText(), "'")[1])
	}
	images := page.MustElementsX("//*[@id=\"productUI\"]/div/div/img/@src")
	for _, image := range images {
		img = append(img, baseUrl+image.MustText())
	}
	fmt.Println("motorList:    ", len(motorList))
	var left string
	leftList := make(map[string]bool)
	comm := map[string]string{
		"Max Torque (N.m)":                  "Max Torque (Nm)",
		"Nominal Voltage (Vdc)":             "Rated Voltage (DC)",
		"Pedal Sensor":                      "Sensor",
		"IP":                                "Tests",
		"Brake":                             "Brake Type",
		"Cabling Route":                     "Cable Routing",
		"Cable Length (mm),Connection Type": "Cable Length (mm)",
		"Weight (kg)":                       "Weight(kg)",
		"Spoke Specification":               "Spoke Spectification",
	}
	var (
		left_l []string
		list_r []string
	)
	for m, motor := range motorList {
		page = broswer.MustPage(motor)
		page.MustWaitLoad()
		motorName := page.MustElementX("//*[@id=\"breadcrumbcontent\"]/ul/li[4]").MustText()

		lists := page.MustElementsX("/html/body/main/div/div[1]/div[2]/div[2]/div[1]/div/p")
		for _, list := range lists {
			left = list.MustElementX("i").MustText()
			right := list.MustElementX("span").MustText()
			if _, ok := leftList[left]; !ok {
				leftList[left] = true
				if _, ok := comm[left]; ok {
					left = comm[left]
				}
				if left == "Position" {
					continue
				}
				left_l = append(left_l, left)
				list_r = append(list_r, right)
			} else {
				continue
			}
		}
		var list_l_str string
		for _, item := range left_l {
			list_l_str = list_l_str + "`" + item + "`,"
		}
		//去掉最后一个逗号
		list_l_str = list_l_str[:len(list_l_str)-1]
		// 插入数据
		query := "INSERT INTO " + "motor_Front" + " (`name`, `url`, `image`, " + list_l_str + ") VALUES (?, ?, ?, " + strings.Repeat("?, ", len(list_r)-1) + "?)"
		// 防注入
		stmt, err := db.Prepare(query)
		if err != nil {
			panic(err)
		}
		defer stmt.Close()
		params := []interface{}{motorName, motor, img[m]}
		for _, item := range list_r {
			params = append(params, item)
		}
		_, err = stmt.Exec(params...)
		if err != nil {
			panic(err)
		}

	}
	fmt.Println("leftList:    ", len(leftList))
	for key := range leftList {
		fmt.Println(key)
	}
}
