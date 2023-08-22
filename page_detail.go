package main

import (
	"database/sql"
	"github.com/go-rod/rod"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"sync"
)

func Details_get(m *sync.Map, lock *sync.Mutex, table string, db *sql.DB, broswer *rod.Browser, url string, works <-chan int, wg *sync.WaitGroup) {
	page := broswer.MustPage(url)
	page.MustWaitDOMStable()
	defer page.MustClose()
	page.WaitElementsMoreThan("#tabs > div.highlights_div > div", 1)
	buttons := page.MustElements("#tabs > div.highlights_div > div.see_more_btn.show_more_highlights")
	if len(buttons) != 0 {
		page.MustElement("#tabs > div.highlights_div > div.see_more_btn.show_more_highlights").MustClick()
		page.MustWaitDOMStable()
	}
	title := page.MustElement("body > div.page-wrapper > section > div.main-content.review-cont-wrapper-box > div > div.left-side-content > h1").MustText()
	img_link := page.MustElementX("/html/body/div[1]/section/div[2]/div/div[1]/div[2]/div[1]/div/img/@src").MustText()
	highlight := page.MustElement("#tabs > div.highlights_div").MustText()
	t_l := page.MustElements("#specs > div > div > div > div.acc_panel > table > tbody > tr > td:nth-child(1)> div > h5")
	lens := len(t_l)
	list_l := make([]string, lens)
	list_r := make([]string, lens)
	for i, Technical := range t_l {
		list_l[i] = Technical.MustText()
		//把空格替换成下划线
		list_l[i] = strings.Replace(list_l[i], " ", "_", -1)
	} //表格的左边
	Ratings := page.MustElements("#specs > div > div > div > div.acc_panel > table > tbody > tr > td:nth-child(2) > div > p")
	for i, Rating := range Ratings {
		list_r[i] = Rating.MustText()
	} //表格的右边
	//判断列是否存在，不存在就创建
	lock.Lock()
	query_l := "ALTER TABLE " + table
	addedColumns := map[string]bool{}
	i := 0
	le := len(list_l)
	for i < le {
		v := list_l[i]
		if addedColumns[v] {
			// 如果这个列名已经添加过，就跳过
			//删除list_l和list_r中的重复元素
			list_l = append(list_l[:i], list_l[i+1:]...)
			list_r = append(list_r[:i], list_r[i+1:]...)
			//fmt.Println("删除重复元素:", v)
			le--
			continue
		}
		addedColumns[v] = true
		if _, ok := m.Load(v); ok {
			i++
			continue
		}
		m.Store(v, true)
		// 如果不存在，创建列
		query_l += " ADD COLUMN `" + v + "` text,"
		//fmt.Println("创建列:", v)
		i++
	}
	if query_l != "ALTER TABLE "+table {
		query_l = query_l[:len(query_l)-1] //去掉最后一个逗号
		query_l += ";"                     //添加分号
		//fmt.Println("query_l:", query_l)
		_, err := db.Exec(query_l)
		if err != nil {
			panic(err)
		}
	}
	lock.Unlock()
	Written := page.MustElement("#ebr_written_rev_inner").MustText()
	exists, ele, err := page.Has("#comments > div > div.comments_inner > div.replied_comments")
	if err != nil {
		panic(err)
	}
	var comment string
	if exists {
		comment = ele.MustText()
	}
	var list_l_str string
	for _, item := range list_l {
		list_l_str = list_l_str + "`" + item + "`,"
	}
	//去掉最后一个逗号
	list_l_str = list_l_str[:len(list_l_str)-1]
	//插入数据
	query := "INSERT INTO " + table + " (`name`, `image`, `Highlights`, `Written_Review`, `Comments`, " + list_l_str + ") VALUES (?, ?, ?, ?, ?, " + strings.Repeat("?, ", len(list_r)-1) + "?)"
	// 创建预编译语句
	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	params := []interface{}{title, img_link, highlight, Written, comment}
	for _, value := range list_r {
		params = append(params, value)
	}
	// 执行预编译语句，传入参数
	_, err = stmt.Exec(params...)
	if err != nil {
		panic(err)
	}
	defer wg.Done()
	defer func() { <-works }()
}
