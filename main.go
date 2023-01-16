package main

import (
	"github.com/saint88/audit_data/pgk/io"
	"strings"

	"flag"
	"fmt"
	"github.com/elliotchance/orderedmap"
	"github.com/saint88/audit_data/pgk/mytracker"
	"github.com/saint88/audit_data/pgk/top"
	"os"
	"strconv"
	"time"
)

// Replace credentials to actual before run
var creds = &mytracker.UserCreds{
	APIUserId: "11111",
	SecretKey: "22222222222222222",
}

var countriesDic = map[string]string{
	"ru": "Россия",
	"am": "Армения",
	"by": "Белоруссия",
	"kz": "Казахстан",
	"kg": "Киргизия",
}

type app struct {
	Type   string
	SDKKey string
	Files  map[string][]*mytracker.File
	Stat   map[string]int
	Header string
}

var applications = map[string][]*app{
	"news": {
		{
			Type:   "Android",
			SDKKey: "06906500589285095554",
			Header: "Новости Mail.ru",
		},
		{
			Type:   "IOS",
			SDKKey: "04793432430918284626",
			Header: "Новости Mail.ru",
		},
	},
	"newsV2": {
		{
			Type:   "Android",
			SDKKey: "59274987045419690955",
			Header: "Новости Mail.ru",
		},
		{
			Type:   "IOS",
			SDKKey: "27974421725893446250",
			Header: "Новости Mail.ru",
		},
	},
	"pharma": {
		{
			Type:   "Android",
			SDKKey: "14780298541803532537",
			Header: "Все Аптеки: Поиск лекарств онлайн",
		},
		{
			Type:   "IOS",
			SDKKey: "84794960176561051381",
			Header: "Все Аптеки: Поиск лекарств!",
		},
	},
}

func main() {
	fmt.Println("====== Сборщик аудиторных показателей приложений медиапроектов ======")
	trackerOnly := flag.Bool("tracker-only", false, "Получаем данные только из https://tracker.my.org")
	rfOnly := flag.Bool("rf-only", false, "Получаем данные только по России")

	appName := flag.String("app", "",
		"Имя приложения по которому будет собираться статистика. Параметр Обязательный.")
	help := flag.Bool("help", false, "Помощь по работе со скриптом")

	year := flag.Int("year", time.Now().Year()-1, "Год за который нужно собрать аудиторские метрики")

	flag.Parse()

	if *appName == "" {
		flag.Usage()
		fmt.Println("Ошибка: Имя приложения для которого собирается статистика не указано")

		fmt.Println("\nСписок поддерживаемых приложений:")
		for name := range applications {
			fmt.Println("-> " + name)
		}

		os.Exit(1)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	var top = make(map[string]int)
	totalTop := 0
	if !*trackerOnly {
		top, totalTop = getDataFromTop(rfOnly, *appName)
	}

	var apps []*app
	for _, app := range getDataFromTracker(rfOnly, year, *appName) {
		stat := make(map[string]int)
		for r, f := range app.Files {
			val := 0
			for _, v := range f {
				msgTemplate := "Загружаем готовый отчет с https://tracker.my.ru о приложении %s для региона %s и платформы %s"
				fmt.Println(fmt.Sprintf(msgTemplate, app.Header, countriesDic[r], app.Type))
				fmt.Println("Download from", v.Link)

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
		fmt.Println(fmt.Sprintf("Статистика https://tracker.my.ru для приложение %s на платформе %s:", app.Header, app.Type))
		for k, v := range app.Stat {
			fmt.Println(fmt.Sprintf("%s: %d", countriesDic[k], v))

			total += v
		}

		fmt.Println()

		totalAllPlatforms += total

		fmt.Println(fmt.Sprintf("Общее количество активности из https://tracker.my.ru для приложение %s на платформе %s: %d", app.Header, app.Type, total))

		// Если с top.mail.ru
		if !*trackerOnly {
			allRegions := 0
			for k, v := range app.Stat {
				allRegions += top[k] + v
			}
			fmt.Println(fmt.Sprintf("Активности со всех источников (https://top.mail.ru + https://tracker.my.ru) для приложение %s на платформе %s: %d", app.Header, app.Type, allRegions))
		}

		fmt.Println()
		fmt.Println()
	}

	fmt.Println()
	fmt.Println(fmt.Sprintf("Всего активности на всех платформах из https://tracker.my.ru: %d", totalAllPlatforms))
	if !*trackerOnly {
		fmt.Println(fmt.Sprintf("Всего активности из всех источников (https://top.mail.ru + https://tracker.my.ru) на всех платформах: %d", totalAllPlatforms+totalTop))
	}
	fmt.Println()
	fmt.Println("Готово!")
}

func getDataFromTracker(rfOnly *bool, year *int, appName string) []*app {
	countries := map[string]int{"ru": 188}
	if !*rfOnly {
		countries["kz"] = 28
		countries["am"] = 215
		countries["by"] = 201
	}

	loc, _ := time.LoadLocation("Europe/Moscow")

	fromDate := time.Date(*year, 1, 1, 0, 0, 0, 0, loc)
	toDate := time.Date(*year, 12, 31, 23, 59, 59, 0, loc)

	var apps []*app
	fmt.Println()

	for _, app := range applications[appName] {
		reportFiles := make(map[string][]*mytracker.File)
		fmt.Println(fmt.Sprintf("Собираем аудиторские данные для приложения %s для %s за %d год", app.Header, app.Type, *year))
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

func getDataFromTop(rfOnly *bool, appName string) (map[string]int, int) {
	top := top.GetAuditMetrics()

	total := 0
	if *rfOnly {
		fmt.Println("Получаем аудиторные показатели из сервиса https://top.mail.ru для России")
		fmt.Println(fmt.Sprintf("%s: %d", countriesDic["ru"], top["ru"]))
		total += top["ru"]
	} else {
		fmt.Println("Получаем аудиторные показатели из сервиса https://top.mail.ru для России")
		for k := range countriesDic {
			fmt.Println(fmt.Sprintf("%s: %d", countriesDic[k], top[k]))
			total += top[k]
		}
	}

	fmt.Println(fmt.Sprintf("Общее количество активности по всем странам для приложения %s: %d", appName, total))

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
	params = append(params, fmt.Sprintf("selectors=idDevice"))

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
