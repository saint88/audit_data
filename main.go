package main

import (
	"AuditNews/src/io"
	"strings"

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
	APIUserId: "1111111",
	SecretKey: "22222222",
}

var countriesDic = map[string]string {
	"ru": "Россия",
	"am": "Армения",
	"by": "Белоруссия",
	"kz": "Казахстан",
	"kg": "Киргизия",
}

type app struct {
	Type string
	SDKKey string
	Files map[string][]*mytracker.File
	Stat map[string]int
}

var applications = []*app{{
	Type:   "Android",
	SDKKey: "06906500589285095554",

}, {
	Type:   "IOS",
	SDKKey: "04793432430918284626",
}}

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

	var top = make(map[string]int)
	totalTop := 0
	if !*trackerOnly {
		top, totalTop = getDataFromTop(rfOnly)
	}

	var apps []*app
	for _, app := range getDataFromTracker(rfOnly) {
		stat := make(map[string]int)
		for r, f := range app.Files {
			val := 0
			for _, v := range f {
				msgTemplate := "Загружаем готовый отчет с https://tracker.my.ru о приложении Новости Mail.ru для региона %s и платформы %s"
				fmt.Println(fmt.Sprintf(msgTemplate, countriesDic[r], app.Type))

				path := fmt.Sprintf("./%s_%s.csv.gz", strings.ToLower(app.Type), r)
				err := io.DownloadFile(path, v.Link)
				if err != nil {
					fmt.Println(err.Error())
				}

				val += io.GetAuditStatFromArchive(path)
			}
			stat[r] = val
		}
		fmt.Println()
		app.Stat = stat

		apps = append(apps, app)
	}

	fmt.Println("Отчет из https://tracker.my.ru:")
	fmt.Println()
	totalAllPlatforms := 0
	for _, app := range apps {
		total := 0
		fmt.Println(fmt.Sprintf("Статистика для приложение Новости Mail.ru на платформе %s:", app.Type))
		for k, v := range app.Stat {
			fmt.Println(fmt.Sprintf("%s: %d", countriesDic[k], v))

			total += v
		}

		fmt.Println()

		allRegions := 0
		for k, v := range app.Stat {
			allRegions = top[k] + v

			total += v
		}

		totalAllPlatforms += total
		fmt.Println(fmt.Sprintf("Общее количество активности из https://tracker.my.ru для приложение Новости Mail.ru на платформе %s: %d", app.Type, total))
		fmt.Println(fmt.Sprintf("Активности со всех источников для приложение Новости Mail.ru на платформе %s: %d", app.Type, allRegions))
	}

	fmt.Println()
	fmt.Println(fmt.Sprintf("Всего активности на всех платформах из https://tracker.my.ru: %d", totalAllPlatforms))
	fmt.Println(fmt.Sprintf("Всего активности из всех источников на всех платформах: %d", totalAllPlatforms + totalTop))
	fmt.Println()
	fmt.Println("Готово!")
}

func getDataFromTracker(rfOnly *bool) []*app {
	countries := map[string]int{"ru": 188}
	if !*rfOnly {
		countries["kz"] = 28
		countries["am"] = 215
		countries["by"] = 201
	}

	fromDate := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2019, 12, 31, 23, 59, 59, 0, time.UTC)

	var apps []*app
	fmt.Println()
	for _, app := range applications {
		reportFiles := make(map[string][]*mytracker.File)
		fmt.Println(fmt.Sprintf("Собираем аудиторские данные для приложения Новости Mail.ru для %s", app.Type))
		for k, v := range countries {
			idRaw := getCreateDataRequest(fromDate, toDate, v, app.SDKKey).IdRawExport

			fmt.Println("Получаем данные из http://tracker.my.com/ для страны " + countriesDic[k])
			reportFiles[k] = getReportFiles(idRaw)
		}

		app.Files = reportFiles
		apps = append(apps, app)
	}

	return apps
}

func getReportFiles(idRaw string) []*mytracker.File {

	url := &mytracker.Url{
		Method: "GET",
		Url:    "https://tracker.my.com/api/raw/v1/export/get.json",
	}

	url.Params = []string{fmt.Sprintf("idRawExport=%s", idRaw)}

	getReport := &mytracker.GetReport{
		Url:   url,
		Creds: creds,
	}

	attempts := 60
	start := 0
	resp := getReport.Get()

	for start < attempts && resp.Data.Status != "Success!" {
		fmt.Print(fmt.Sprintf("%s ", resp.Data.Progress))
		start++
		resp = getReport.Get()
		time.Sleep(1 * time.Second)
	}
	fmt.Print("100%")
	fmt.Println()
	fmt.Println(fmt.Sprintf("Статус получения данных из http://tracker.my.com/: %s", resp.Data.Status))

	fmt.Println()

	if len(resp.Data.Files) == 0 {
		fmt.Println("Получение данных из http://tracker.my.com/ завершилось ошибкой")
		os.Exit(1)
	}

	return resp.Data.Files
}

func getDataFromTop(rfOnly *bool) (map[string]int, int) {
	top := top.GetAuditMetrics()

	total := 0
	if *rfOnly {
		fmt.Println("Получаем аудиторные показатели из сервиса https://top.mail.ru для России")
		fmt.Println(fmt.Sprintf("%s: %d", countriesDic["ru"], top["ru"]))
		total += top["ru"]
	} else {
		for k := range countriesDic {
			fmt.Println(fmt.Sprintf("%s: %d", countriesDic[k], top[k]))
			total += top[k]
		}
	}

	fmt.Println(fmt.Sprintf("Общее количество активности по всем странам для приложения Новости Mail.Ru: %d", total))

	return top, total
}

func getCreateDataRequest(fromDate time.Time, toDate time.Time, country int, sdkKeyApp string) *mytracker.ReportData {
	var event = "activities"

	m := orderedmap.NewOrderedMap()
	m.Set("tsFrom", strconv.FormatInt(fromDate.Unix(), 10))
	m.Set("tsTo", strconv.FormatInt(toDate.Unix(), 10))

	var params []string
	params = append(params, fmt.Sprintf("SDKKey=%s", sdkKeyApp))
	params = append(params, fmt.Sprintf("event=%s", event))
	params = append(params, fmt.Sprintf("idCountry=%d", country))
	params = append(params, fmt.Sprintf("selectors=dtEvent"))

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
