package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"

	"github.com/Unknwon/goconfig"
	"github.com/otwdev/galaxylib"
)

type DataGeter interface {
	GetData() (ret []*BillItem)
}

type Alihl struct {
	//FileName string
	cfg    *goconfig.ConfigFile
	remote *url.URL
	tr     *goquery.Selection
	token  string
	jar    *cookiejar.Jar
}

type BillItem struct {
	CreationTime         string
	TradeNo              string
	OrderNumber          string
	OtherAccount         string
	OtherName            string
	Airline              string
	ProductName          string
	PNR                  string
	Category             string
	InCome               float64
	Expeness             float64
	Balance              float64
	CreditLimit          string
	CreditPurchaser      string
	StartingTktNumber    string
	EndTicketNo          string
	AccountFlowNo        string  //账务流水号
	Memo                 string  //备注
	TotalAvailableAmount float64 // 总账余额
	Account              string
}

//var jar *cookiejar.Jar

func NewAlihl(filename string) *Alihl {
	a := &Alihl{}
	//a.FileName = filename

	a.cfg, _ = goconfig.LoadConfigFile(filename)
	a.remote, _ = url.Parse("https://hl.alipay.com/bill/2005/200501/product.htm")
	a.loadCookie()
	return a
}

func (a *Alihl) loadCookie() {

	if a.jar != nil {
		return
	}

	cookies := a.cfg.MustValueArray("var", "cookie", ";")

	var cookieCollection []*http.Cookie
	for _, v := range cookies {
		kv := strings.Split(v, "=")
		cookie := &http.Cookie{
			Name:  strings.TrimSpace(kv[0]),
			Value: strings.TrimSpace(kv[1]),
		}
		cookieCollection = append(cookieCollection, cookie)
	}
	a.jar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	//}) //galaxylib.NewJar()
	a.jar.SetCookies(a.remote, cookieCollection)
}

func (a *Alihl) GetData() (ret []*BillItem) {

	loadBuf := a.request("GET", nil)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(loadBuf))
	if err != nil {
		fmt.Println(err)
		return
	}

	if a.token == "" {
		a.token, _ = doc.Selection.Find("input[name=_form_token]").Attr("value")
	}
	lastweek, _ := doc.Selection.Find("input[name=lastWeek]").Attr("value")
	toDayDate, _ := doc.Selection.Find("input[name=todayDate]").Attr("value")

	time.Sleep(2 * time.Second)

	accounts, _ := a.cfg.GetSection("account")

	for k, v := range accounts {
		rev := a.postQuery(lastweek, toDayDate, k, v)
		ret = append(ret, rev...)
		time.Sleep(1 * time.Second)
	}

	return

	//a.postQuery(lastweek, toDayDate, "2088501659368760")

	//query := "=%s&=0&=1&=40&=&=&=%s&=%s&=&=2018-07-06&=2018-08-06&=&=&=&=&=&=&="

	//fmt.Println(token)
}

func (a *Alihl) postQuery(lastweek, toDayDate, cardNo, account string) (revData []*BillItem) {

	query := url.Values{}

	query.Add("_form_token", a.token)
	query.Add("showDetail", "0")
	query.Add("pageStr", "1")
	query.Add("pageSize", "40")
	query.Add("fold", "false")
	query.Add("collectionHeighSearchId", "")
	query.Add("lastWeek", lastweek)
	query.Add("todayDate", toDayDate)
	query.Add("createTimeOrderType", "TRANS_DT_DESC")
	query.Add("beginDate", toDayDate)
	query.Add("endDate", toDayDate)
	query.Add("cardNo", cardNo)
	query.Add("outTradeNo", "")
	query.Add("partnerId", "")
	query.Add("airBizType", "")
	query.Add("pnr", "")
	query.Add("ticketNo", "")
	query.Add("incomeOutcome", "")

	params := strings.NewReader(query.Encode())

	fmt.Println(query.Encode())

	postBuf := a.request("POST", params)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(postBuf))

	if err != nil {
		fmt.Println(err)
		return
	}

	html, _ := doc.Html()

	html = galaxylib.DefaultGalaxyTools.Bytes2CHString([]byte(html))

	//fmt.Println(html)

	doc.Selection.Find("#JSServiceTable>tbody>tr").Each(func(index int, slt *goquery.Selection) {

		//fmt.Println(index)

		item := &BillItem{}
		a.tr = slt
		item.Account = account
		item.CreationTime = a.getTdData("createTime")
		item.TradeNo = a.getTdData("tradeNo", true)
		item.OtherAccount = a.getTdData("otherEmail")
		item.OtherName = a.getTdData("otherName")
		item.Airline = a.getTdData("platformName")
		item.ProductName = a.getTdData("goodsTitle")
		item.PNR = a.getTdData("pnr")
		item.Category = a.getTdData("airBizType")
		item.InCome = galaxylib.DefaultGalaxyConverter.MustFloat(a.getTdData("inAmount"))
		item.Expeness = galaxylib.DefaultGalaxyConverter.MustFloat(a.getTdData("outAmount"))
		item.Balance = galaxylib.DefaultGalaxyConverter.MustFloat(a.getTdData("balance"))
		item.CreditLimit = a.getTdData("creditAvailableAmount")
		item.CreditPurchaser = a.getTdData("transOperator")
		item.StartingTktNumber = a.getTdData("startTicketNo")
		item.EndTicketNo = a.getTdData("endTicketNo")
		item.AccountFlowNo = a.getTdData("accountFlowNo")
		item.Memo = a.getTdData("memo")
		item.TotalAvailableAmount = galaxylib.DefaultGalaxyConverter.MustFloat(a.getTdData("totalAvailableAmount"))
		revData = append(revData, item)

		fmt.Printf("%s---%f\n", item.TradeNo, item.Balance)
	})

	a.token, _ = doc.Selection.Find("input[name=_form_token]").Attr("value")

	// h, _ := tbl.Html()
	// fmt.Println(h)

	//tbl.FindMatcher

	// html, err := doc.Html()

	// html = galaxylib.DefaultGalaxyTools.Bytes2CHString([]byte(html))

	// //fmt.Println(html)
	// revData = nil
	return
}

func (a *Alihl) getTdData(name string, isLink ...bool) string {
	selector := fmt.Sprintf("td[colName=%s]>cite", name)
	if len(isLink) > 0 {
		selector = fmt.Sprintf("%s>a", selector)
	}
	val := a.tr.Find(selector).Text()
	return strings.TrimSpace(val)
}

func (a *Alihl) request(method string, body io.Reader) []byte {

	rq, err := http.NewRequest(method, a.remote.String(), body)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	heaerAry, _ := galaxylib.GalaxyCfgFile.GetSection("header")

	for k, v := range heaerAry {
		rq.Header.Add(k, v)
	}
	if method == "POST" {
		rq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	client := http.DefaultClient
	client.Jar = a.jar

	rs, err := client.Do(rq)

	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rs.Body.Close()

	// for c, _ := range rs.Cookies() {
	// 	a.jar.SetCookies(rq.URL, c)
	// }

	buf, _ := ioutil.ReadAll(rs.Body)
	//fmt.Println(galaxylib.DefaultGalaxyTools.Bytes2CHString(buf))
	return buf

}
