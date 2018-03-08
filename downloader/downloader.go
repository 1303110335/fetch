package downloader

import (
	"net/http"
	mdw "sys/fetch/middleware"
	"sys/fetch/logging"
	base "sys/fetch/base"
)

// 日志记录器。
var logger logging.Logger = base.NewLogger()

type PageDownloader interface {
	Id() uint32 //获得ID
	Download(req base.Request) (*base.Response, error) //根据请求下载网页并返回响应
}

//ID生成器接口类型
type IdGenertor interface {
	GetUint32() uint32 //获得一个uint32类型的ID
}

//网页下载器池的接口类型
type PageDownloaderPool interface {
	Take() (PageDownloader, error)  //从池中取出一个网页下载器
	Return(dl PageDownloader) error //把一个网页下载器归还给池
	Total() uint32					//获得池的总量
	Used() uint32					//获得正在被使用的网页下载器的数量
}

type myPageDownloader struct {
	httpClient http.Client	//Http客户端
	id 			uint32		//ID
}

//ID生成器
var downloaderIdGenertor mdw.IdGenerator = mdw.NewIdGenerator()

//生成并返回ID
func genDownloaderId() uint32 {
	return downloaderIdGenertor.GetUint32()
}

// 创建网页下载器。
func NewPageDownloader(client *http.Client) PageDownloader {
	id := genDownloaderId()
	if client == nil {
		client = &http.Client{}
	}
	return &myPageDownloader{
		id:         id,
		httpClient: *client,
	}
}

func (dl *myPageDownloader) Id() uint32 {
	return dl.id
}

func (dl *myPageDownloader) Download(req base.Request) (*base.Response, error) {
	httpReq := req.HttpReq()
	logger.Infof("Do the request (url=%s)... \n", httpReq.URL)
	httpResp, err := dl.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	return base.NewResponse(httpResp, req.Depth()), nil
}
