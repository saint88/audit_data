package top

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func check(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

func GetAuditMetrics() map[string]int64 {
	resp, err := http.Get("https://top.mail.ru/74867-2019-CC.txt")

	defer resp.Body.Close()

	check(err)

	b, err := ioutil.ReadAll(resp.Body)
	check(err)

	auditInfo := make(map[string] int64)

	metrics := strings.Split(strings.TrimSpace(string(b)), "\n")
	for _, metric := range metrics {
		metricInRegion := strings.Split(metric, "\t")

		m, err := strconv.ParseInt(metricInRegion[1], 10, 64)
		check(err)

		auditInfo[strings.ToLower(metricInRegion[0])] = m
	}

	return auditInfo
}
