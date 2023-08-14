package main

import (
	"bytes"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/xuri/excelize/v2"
	"golang.org/x/image/webp"
	"image/jpeg"
	"io"
	"net/http"
	"sync"
)

func convertIntToLetters(i int) string {

	if i < 0 {
		return ""
	}

	s := ""

	for i > 0 {
		m := i % 26
		if m == 0 {
			m = 26
		}
		s = string('A'+m-1) + s
		i = (i - m) / 26
	}

	return s
}
func Details_get(sv int, lock *sync.Mutex, i string, sheet string, f *excelize.File, broswer *rod.Browser, url string, works <-chan int, wg *sync.WaitGroup) {
	page := broswer.MustPage(url)
	page.MustWaitDOMStable()
	defer page.MustClose()
	page.WaitElementsMoreThan("#tabs > div.highlights_div > div", 1)
	buttons := page.MustElements("#tabs > div.highlights_div > div.see_more_btn.show_more_highlights")
	if len(buttons) != 0 {
		page.MustElement("#tabs > div.highlights_div > div.see_more_btn.show_more_highlights").MustClick()
		page.MustWaitDOMStable()
	}
	//fmt.Println("================================================")
	lock.Lock()
	title := page.MustElement("body > div.page-wrapper > section > div.main-content.review-cont-wrapper-box > div > div.left-side-content > h1")
	//fmt.Println(title.MustText()) //标题
	//fmt.Println("========================================================", "A"+i)
	f.SetCellValue(sheet, "A1", "name")
	f.SetCellValue(sheet, "A"+i, title.MustText())
	img := page.MustElementX("/html/body/div[1]/section/div[2]/div/div[1]/div[2]/div[1]/div/img/@src")
	//fmt.Println(img.MustText()) //图片链接
	resp, err := http.Get(img.MustText())
	if err != nil {
		panic(err)
	}
	image_webp, err := io.ReadAll(resp.Body)
	if err != nil {
		image_webp = nil
	}
	img_coder, err := webp.Decode(bytes.NewReader(image_webp))
	if err != nil {
		panic(err)
	}
	img_jpg := new(bytes.Buffer)
	err = jpeg.Encode(img_jpg, img_coder, nil)
	if err != nil {
		panic(err)
	}
	img_data := img_jpg.Bytes()
	f.SetCellValue(sheet, "B1", "image")
	if err := f.AddPictureFromBytes(sheet, "B"+i, &excelize.Picture{
		Extension: ".jpg",
		File:      img_data,
		Format: &excelize.GraphicOptions{AltText: "Excel Logo",
			AutoFit: true,
		},
	}); err != nil {
		panic(err)
	}
	resp.Body.Close()
	//fmt.Println("================================================") //heighlights
	//fmt.Println("Highlights\n", page.MustElement("#tabs > div.highlights_div").MustText())
	f.SetCellValue(sheet, "C1", "Highlights")
	f.SetCellValue(sheet, "C"+i, page.MustElement("#tabs > div.highlights_div").MustText())
	//fmt.Println("================================================")
	Technicals := page.MustElements("#specs > div > div > div > div.acc_panel > table > tbody > tr > td:nth-child(1)> div > h5")
	lens := len(Technicals)
	list := make([]string, lens*2)
	for i, Technical := range Technicals {
		list[2*i] = Technical.MustText()
	} //表格的左边
	//fmt.Println("================================================")
	Ratings := page.MustElements("#specs > div > div > div > div.acc_panel > table > tbody > tr > td:nth-child(2) > div > p")
	for i, Rating := range Ratings {
		list[2*i+1] = Rating.MustText()
	} //表格的右边
	//lock.Lock()
	for k := 0; k < len(list); k += 2 {
		result, err := f.SearchSheet(sheet, list[k])
		if err != nil {
			lock.Unlock()
			fmt.Println(err)
			return
		}
		if len(result) == 0 {
			//lock.Lock()
			err := f.InsertCols(sheet, "F", 1)
			if err != nil {
				lock.Unlock()
				fmt.Println(err)
				return
			}
			//lock.Unlock()
			f.SetCellValue(sheet, "F1", list[k])
			f.SetCellValue(sheet, "F"+i, list[k+1])
		} else {
			f.SetCellValue(sheet, result[0][:len(result[0])-1]+i, list[k+1])
		}
	}
	//lock.Unlock()
	//fmt.Println("================================================")
	Written := page.MustElement("#ebr_written_rev_inner")
	//fmt.Println(Written.MustText()) //Written Review
	f.SetCellValue(sheet, "D1", "Written Review")
	f.SetCellValue(sheet, "D"+i, Written.MustText())
	//fmt.Println("================================================")
	comments := page.MustElements("#comments > div > div.comments_inner > div.replied_comments")
	//fmt.Println(len(comments))
	for _, comment := range comments {
		f.SetCellValue(sheet, "E1", "Comments")
		f.SetCellValue(sheet, "E"+i, comment.MustText())
	}
	if sv == 1 {
		fmt.Println(i, "===>Save")
		f.Save()
	}
	lock.Unlock()
	defer wg.Done()
	defer func() { <-works }()
	defer func() {
		fmt.Println(i, "===>Done")
	}()

}
