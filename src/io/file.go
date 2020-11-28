package io

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func check(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	check(err)

	defer resp.Body.Close()

	out, err := os.Create(filepath)
	check(err)
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	return err
}

func GetAuditStatFromArchive(path string) int {
	f, err := os.Open(path)
	check(err)
	defer f.Close()

	gr, err := gzip.NewReader(f)
	check(err)
	defer gr.Close()

	dat, err := ioutil.ReadAll(gr)
	check(err)

	s := strings.Split(string(dat), "\n")
	defer func() {
		err := os.RemoveAll(path)
		check(err)
	}()

	s = s[1:]
	set := make(map[string]bool)
	for _, id := range s {
		set[id] = true
	}

	return len(set)
}
