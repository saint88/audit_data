package main

import (
	"AuditNews/src/mytracker"
	"AuditNews/src/top"
	"flag"
	"fmt"
	"github.com/elliotchance/orderedmap"
	"os"
	"strconv"
	"time"
)

var creds = &mytracker.UserCreds{
	APIUserId: "11111",
	SecretKey: "12345678901234567890",
}

func main() {
	fmt.Println("====== Сборщик аудиторных показателей приложения Новости Mail.ru Group ======")
	trackerOnly := flag.Bool("tracker-only", false, "Получаем данные только из https://tracker.my.org")
	rfOnly := flag.Bool("rf-only", false, "Получаем данные только по России")
	help := flag.Bool("help", false, "Помощь по работе со скриптом")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if !*trackerOnly {
		getDataFromTop(rfOnly)
	}

	getDataFromTracker(rfOnly)
}

func getDataFromTracker(rfOnly *bool) {
	countries := map[string]int{"ru": 188}
	if !*rfOnly {
		countries["kz"] = 28
		countries["am"] = 215
		countries["by"] = 201
	}

	fromDate := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2019, 12, 31, 23, 59, 59, 0, time.UTC)

	idRaw := getCreateDataRequest(fromDate, toDate, countries).IdRawExport

	url := &mytracker.Url{
		Method: "GET",
		Url:    "https://tracker.my.com/api/raw/v1/export/get.json",
	}

	url.Params = []string{fmt.Sprintf("idRawExport=%s", idRaw)}

	getReport := &mytracker.GetReport{
		Url:   url,
		Creds: creds,
	}

	attempts := 20
	start := 0
	resp := getReport.Get()

	fmt.Println("Status get data from http://tracker.my.com/ " + resp.Data.Status)
	for start < attempts && resp.Data.Status != "Success!" {
		fmt.Print(fmt.Sprintf("%s ", resp.Data.Progress))
		start++
		resp = getReport.Get()
		time.Sleep(1 * time.Second)
	}

	fmt.Println()

	fmt.Println("Get audit info of 'News Mobile Application' on Android and IOS from http://tracker.my.com/ Finished!!!")
	fmt.Println("Link: " + resp.Data.Files[0].Link)
}

func getDataFromTop(rfOnly *bool) {
	regions := make(map[string]string)
	regions["ru"] = "Россия"
	regions["am"] = "Армения"
	regions["by"] = "Белоруссия"
	regions["kz"] = "Казахстан"
	regions["kg"] = "Киргизия"

	top := top.GetAuditMetrics()

	if *rfOnly {
		fmt.Println("Получаем аудиторные показатели из сервиса https://top.mail.ru для России")
		fmt.Println(fmt.Sprintf("%s: %d", regions["ru"], top["ru"]))
	} else {
		for k := range regions {
			fmt.Println(fmt.Sprintf("%s: %d", regions[k], top[k]))
		}
	}
}

func getCreateDataRequest(fromDate time.Time, toDate time.Time, countries map[string]int) *mytracker.ReportData {
	var applications = []string{"06906500589285095554", "04793432430918284626"}
	//var countries = map[string]int{"kz": 28, "ru": 188, "am": 215, "by": 201}
	var selectors = []string{"dtEvent", "SDKKey", "idAppTitle", "idCountryTitle", "idDeviceModelTitle", "idManufacturerTitle"}
	var event = "activities"

	m := orderedmap.NewOrderedMap()
	m.Set("tsFrom", strconv.FormatInt(fromDate.Unix(), 10))
	m.Set("tsTo", strconv.FormatInt(toDate.Unix(), 10))

	var params []string
	for _, sdkKey := range applications {
		params = append(params, fmt.Sprintf("SDKKey=%s", sdkKey))
	}

	params = append(params, fmt.Sprintf("event=%s", event))

	for _, c := range countries {
		params = append(params, fmt.Sprintf("idCountry=%d", c))
	}

	for _, s := range selectors {
		params = append(params, fmt.Sprintf("selectors=%s", s))
	}

	for _, k := range m.Keys() {
		v, _ := m.Get(k)
		s := fmt.Sprintf("%s=%s", k, v)
		params = append(params, s)
	}

	url := &mytracker.Url{
		Method: "POST",
		Url:    "https://tracker.my.com/api/raw/v1/export/create.json",
	}

	url.Params = params

	createDataReqMyTracker := &mytracker.CreateReport{
		Url:   url,
		Creds: creds,
	}

	resp := createDataReqMyTracker.Create()

	return resp.ReportData
}
