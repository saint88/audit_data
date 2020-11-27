package mytracker

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	url2 "net/url"
	"strings"
)

type Application struct {
	SDKKey int64
}

type UserCreds struct {
	APIUserId string
	SecretKey string
}

type Url struct {
	Url string
	Method string
	Params []string
}

type CreateReport struct {
	Url *Url
	Creds *UserCreds
	SDKKey []string
	IdCountry []string
	Event string
	Selectors []string
	DateFrom string
	DateTo string
}

type GetReport struct {
	Url *Url
	Creds *UserCreds
	IdRawExport string
}

type GetResp struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data *ReportData `json:"data"`
}

type CreateResp struct {
	Code int `json:"code"`
	Message string `json:"message"`
	ReportData *ReportData `json:"data"`
}

type ReportData struct {
	IdRawExport string `json:"idRawExport"`
	Status string `json:"status"`
	ErrorMessage string `json:"errorMessage"`
	Progress string `json:"progress"`
	IsCancellable bool `json:"isCancellable"`
	Files []*File `json:"files"`
}

type File struct {
	Link string `json:"link"`
	Timestamp string `json:"timestampExpires"`
}

type MyTracker interface {
	Create()
	Get()
}

func check(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

func createAuthSubscribe(u *Url, creds *UserCreds) string {
	url := u.Url + "?" + strings.Join(u.Params, "&")
	urlEncoded := url2.QueryEscape(url)

	baseString := u.Method + "&" + urlEncoded + "&"

	h := hmac.New(sha1.New, []byte(creds.SecretKey))
	h.Write([]byte(baseString))
	sub := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return "AuthHMAC " + creds.APIUserId + ":" + sub
}

func (createReport *CreateReport) Create() *CreateResp {
	client := http.DefaultClient
	req, err := http.NewRequest(http.MethodPost, createReport.Url.Url, nil)
	check(err)
	req.Header.Add("Authorization", createAuthSubscribe(createReport.Url, createReport.Creds))

	q := req.URL.Query()
	for _, v := range createReport.Url.Params {
		r := strings.Split(v, "=")
		q.Add(r[0], r[1])
	}

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	check(err)

	var createResp *CreateResp
	err = json.Unmarshal(b, &createResp)
	check(err)

	return createResp
}

func (getReport *GetReport) Get() *GetResp {
	client := http.DefaultClient
	req, err := http.NewRequest(http.MethodGet, getReport.Url.Url, nil)
	check(err)

	req.Header.Add("Authorization", createAuthSubscribe(getReport.Url, getReport.Creds))

	q := req.URL.Query()
	for _, v := range getReport.Url.Params {
		r := strings.Split(v, "=")
		q.Add(r[0], r[1])
	}

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	check(err)

	var createResp *GetResp
	err = json.Unmarshal(b, &createResp)
	check(err)

	return createResp
}