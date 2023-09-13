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
	baseUrl := "https://ananda.com.cn/"
	page := broswer.MustPage("https://ananda.com.cn/index.php/chinese/product/index.html")
	db, err := sql.Open("mysql", "root:heanyang@tcp(10.199.1.41:8848)/motor?charset=utf8mb4&parseTime=True")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	defer db.Close()
	page.MustWaitDOMStable()
	category := page.MustElementsX("//*[@id=\"picabout\"]/div[2]/div/div")
	// 分三个类别，前置中置和后置
	for _, c := range category[1:] {
		table := make(map[string]bool)
		////创建表
		name := c.MustElementX("div[1]/h4").MustText()
		fmt.Println(name)
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS " + name + " ( `id` int auto_increment,name char(100) default 'Error', image varchar(800) default 'Error',primary key(id),INDEX name (name)); ")
		if err != nil {
			panic(err)
		}
		motors := c.MustElementsX("div[2]/div/div")
		var (
			motorLinks []string
			link       string
			img        string
		)
		// 获取每个电机的链接和图片
		var (
			list_l []string // 左侧列表
			list_r []string // 右侧列表
		)
		for i, motor := range motors {
			list_l = nil
			list_r = nil
			fmt.Println(i, "=========")
			// list为了去重
			list := make(map[string]int)
			link = baseUrl + motor.MustElementX("a/@href").MustText()
			img = baseUrl + motor.MustElementX("div/div/img/@src").MustText()
			motorLinks = append(motorLinks, link)
			page = broswer.MustPage(link)
			page.MustWaitLoad()
			motorName := page.MustElementX("//*[@id=\"picabout\"]/div[2]/div[1]/div/h1").MustText()
			var (
				left  string
				right string
			)
			query_l := "ALTER TABLE " + name
			for k, itom := range page.MustElementsX("//*[@id=\"picabout\"]/div[2]/div[2]/div/div[2]/ul/li")[3:] {
				left = itom.MustElementX("div[1]").MustText()
				right = itom.MustElementX("div[2]").MustText()
				// 歪比巴卜
				if right == "" {
					fmt.Println("right is nil")
					continue
				}
				// 玛卡巴卡
				if left == "" {
					list_r[len(list_r)-1] = list_r[len(list_r)-1] + "," + right
					continue
				}
				if list[left] == 0 {
					list[left] = k
					list_l = append(list_l, left)
					list_r = append(list_r, right)
					if !table[left] {
						query_l += " ADD COLUMN `" + left + "` text,"
						table[left] = true
					}
				} else {
					fmt.Println(left, "已存在")
					list_r[list[left]] = list_r[list[left]] + "," + right
				}
			}
			//if query_l != "ALTER TABLE "+name {
			if query_l != "ALTER TABLE "+name {
				query_l = query_l[:len(query_l)-1] //去掉最后一个逗号
				query_l += ";"                     //添加分号
				_, err := db.Exec(query_l)         //执行sql语句,创建列
				if err != nil {
					panic(err)
				}
			}
			// 构造插入语句
			var list_l_str string
			for _, item := range list_l {
				list_l_str = list_l_str + "`" + item + "`,"
			}
			//去掉最后一个逗号
			list_l_str = list_l_str[:len(list_l_str)-1]
			// 插入数据
			//fmt.Println(list_l_str)
			query := "INSERT INTO " + name + " (`name`, `image`, " + list_l_str + ") VALUES (?, ?, " + strings.Repeat("?, ", len(list_r)-1) + "?)"
			// 防注入
			stmt, err := db.Prepare(query)
			if err != nil {
				panic(err)
			}
			defer stmt.Close()
			params := []interface{}{motorName, img}
			for _, item := range list_r {
				params = append(params, item)
			}
			_, err = stmt.Exec(params...)
			if err != nil {
				panic(err)
			}
		}
	}
}
