package main

import (
	"sys/fetch/base"
	"net/http"
	"sys/fetch/logging"
	"errors"
	"fmt"
	"net/url"
	"io"
	"github.com/Puerkitobio/goquery"
	"strings"
	"sys/fetch/analyzer"
	"time"
	pipeline "sys/fetch/itempipeline"
	sched "sys/fetch/scheduler"
)

var logger logging.Logger = logging.NewSimpleLogger()

func main() {
	//创建调度器
	scheduler := sched.NewScheduler()

	//准备启动参数
	channelArgs := base.NewChannelArgs(10, 10, 10, 10)
	poolBaseArgs := base.NewPoolBaseArgs(3, 3)
	crawlDepth := uint32(1)
	httpClientGenerator := getHttpClient
	respParsers := getResponseParsers()
	itemProcessors := getItemProcessors()
	startUrl := "http://www.sougou.com"
	firstHttpReq, err := http.NewRequest("GET", startUrl, nil)
	if err != nil {
		logger.Errorln(err)
		return
	}

	//开启调度器
	scheduler.Start(
		channelArgs,
		poolBaseArgs,
		crawlDepth,
		httpClientGenerator,
		respParsers,
		itemProcessors,
		firstHttpReq)

}

func getHttpClient() *http.Client {
	return &http.Client{}
}

//获得响应解析函数的序列
func getResponseParsers() []analyzer.ParseResponse {
	parsers := []analyzer.ParseResponse{
		parseForATag,
	}
	return parsers
}

//条目处理器
func processItem(item base.Item) (result base.Item, err error) {
	if item == nil {
		return nil, errors.New("Invalid item!")
	}
	//生成结果
	result = make(map[string]interface{})
	for k, v := range item {
		result[k] = v
	}
	if _, ok := result["number"]; !ok {
		result["number"] = len(result)
	}
	time.Sleep(10 * time.Millisecond)
	return result, nil
}

// 获得条目处理器的序列
func getItemProcessors() []pipeline.ProcessItem {
	itemProcessors := []pipeline.ProcessItem {
		processItem,
	}
	return itemProcessors
}

//响应解析函数。只解析"A"标签
func parseForATag(httpResp *http.Response, respDepth uint32) ([]base.Data, []error) {
	// TODO 支持更多的HTTP响应
	if httpResp.StatusCode != 200 {
		err := errors.New(
			fmt.Sprintf("Unsupported status code %d.(httpResponse=%v)", httpResp))
			return nil, []error{err}
	}

	var reqUrl *url.URL = http.Request.URL
	var httpRespBody io.ReadCloser = httpResp.Body
	defer func() {
		if httpRespBody != nil {
			httpRespBody.Close()
		}
	}()
	dataList := make([]base.Data, 0)
	errs := make([]error, 0)

	//开始解析
	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		errs = append(errs, err)
		return dataList, errs
	}

	// 查找“A“标签并提取链接地址
	doc.Find("a").Each(func(index int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		//前期过滤
		if !exists || href == "" || href == "#" || href == "/" {
			return
		}
		href = strings.TrimSpace(href)
		lowerHref := strings.ToLower(href)
		//暂不支持对javascript代码的解析
		if href != "" && !strings.HasPrefix(lowerHref, "javascript") {
			aUrl, err := url.Parse(href)
			if err != nil {
				errs = append(errs, err)
				return
			}
			if !aUrl.IsAbs() {
				aUrl = reqUrl.ResolveReference(aUrl)
			}
			httpReq, err := http.NewRequest("GET", aUrl.String(), nil)
			if err != nil {
				errs = append(errs, err)
			} else {
				req := base.NewRequest(httpReq, respDepth)
				dataList = append(dataList, req)
			}
		}
		text := strings.TrimSpace(sel.Text())
		if text != "" {
			imap := make(map[string]interface{})
			imap["a.text"] = text
			imap["parent_url"] = reqUrl
			item := base.Item(imap)
			dataList = append(dataList, &item)
		}
	})
	return dataList, errs
}