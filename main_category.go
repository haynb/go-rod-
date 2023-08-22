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
	for _, category_link := range categories_links {
		name := category_link[40 : len(category_link)-1]
		if len(name) > 30 {
			name = name[:30]
		}
		name = strings.ReplaceAll(name, "-", "_")
		fmt.Println("gogogo!!!!!!!!!!!!!!!!!!!!!!!!")
		//连接数据库
		//db, err := sql.Open("mysql", "root:heanyang@tcp(154.12.244.129:8848)/Category?charset=utf8mb4&parseTime=True")
		db, err := sql.Open("mysql", "root:heanyang@tcp(10.199.1.41:8848)/Category?charset=utf8mb4&parseTime=True")
		if err != nil {
			panic(err)
		}
		defer db.Close()
		db.SetMaxOpenConns(20)
		db.SetMaxIdleConns(10)
		fmt.Println("name:", name)
		rows, err := db.Query("SHOW TABLES LIKE " + "'" + name + "'" + ";")
		if err != nil {
			panic(err)
		}
		if rows.Next() {
			fmt.Println("表格存在，执行 continue")
			rows.Close()
			continue
		}
		rows.Close()
		//_, err = db.Exec("DROP TABLE IF EXISTS " + name)
		//if err != nil {
		//	panic(err)
		//}
		//创建表
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS " + name + " ( `id` int auto_increment,name char(100) default 'Error', image varchar(800) default 'Error',Highlights text, Written_Review text,Comments text,primary key(id),INDEX name (name)); ")
		if err != nil {
			panic(err)
		}
		Go_details(name, db, broswer, category_link)
		fmt.Println("a category done!!!!!!!!!!!!!!!!!!!!!!!!")
	}
}
