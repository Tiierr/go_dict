package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	a "github.com/logrusorgru/aurora"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func getUrl(q string) string {
	baseUrl, _ := url.Parse("http://fanyi.youdao.com/openapi.do")
	params := url.Values{}

	payload := map[string]string{
		"keyfrom": "wufeifei",
		"key": "716426270",
		"type": "data",
		"doctype": "json",
		"version": "1.1",
		"q": q,
	}
	for k := range payload{
		params.Set(k, payload[k])
	}

	baseUrl.RawQuery = params.Encode()
	fullUrl := baseUrl.String()
	return fullUrl
}

func getQueryData(url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Fatal(err)
	}
	return body
}


func main()  {
	args := os.Args[1:]
	strArgs := strings.Join(args, " ")

	body := getQueryData(getUrl(strArgs))

	//var result Result
	result, err := simplejson.NewJson(body)
	if err != nil {
		log.Fatal(err)
	}


	// handle error code
	errorCode := result.Get("errorCode").MustInt()
	var errorMsg string
	switch errorCode {
	case 0:
		errorMsg = ""
	case 20:
		errorMsg = "要翻译的文本过长"
	case 30:
		errorMsg = "无法进行有效的翻译"
	case 40:
		errorMsg = "不支持的语言类型"
	case 50:
		errorMsg = "无效的 key"
	case 60:
		errorMsg = "无词典结果"
	}

	if errorMsg != "" {
		fmt.Println(a.Bold(a.Red(errorMsg)))
		return
	}

	// 1. query and phonetic detail
	basic, _ := result.Get("basic").Map()
	translation, _ := result.Get("translation").Array()
	query, _ := result.Get("query").String()

	var queryStr, phoneticStr, fillStr string
	queryStr = fmt.Sprintf("%s  %s", a.Cyan("⠸"), query)

	if basic["phonetic"] != nil{
		phoneticStr = fmt.Sprintf("  [ %s ]", basic["phonetic"])
	}

	fillStr = "  ~  "
	fmt.Printf("\n%s%s%s%s\n\n", queryStr, a.Magenta(phoneticStr), fillStr, a.Bold(a.Gray(25, translation[0])))

	// 2. common explains
	for _, explains :=range result.Get("basic").Get("explains").MustArray() {
		prefix := " - "
		fmt.Printf("%s%s\n", prefix, a.Bold(a.Green(explains)))
	}

	// 3. web explains
	fmt.Println()
	for i, web :=range result.Get("web").MustArray() {
		webInterface := web.(map[string]interface{})
		key := webInterface["key"]
		value := webInterface["value"]

		// 3.1 示例
		prefix := ". "
		fmt.Printf(" %d%s%s\n", a.Gray(15, i+1), prefix, a.Yellow(key))

		// 3.2 示例翻译
		fourSpace := strings.Repeat(" ", 4)
		aInterface := value.([]interface{})
		aString := make([]string, len(aInterface))
		for i, v := range aInterface {
			aString[i] = v.(string)
		}

		oneLineString := strings.Join(aString[:],",")
		fmt.Printf("%s%s\n", fourSpace, a.Cyan(oneLineString))

	}
}