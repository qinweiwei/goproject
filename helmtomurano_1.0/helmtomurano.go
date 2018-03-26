package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var listfile []string //获取文件列表

type ChartConfig struct {
	Name     string `json:"name"`
	Describe string `json:"describe"`
	Default  string `json:"default"`
}

type HelmChart struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Repo    string `json:"repo"`
	Config  []ChartConfig
}

func Listfunc(path string, f os.FileInfo, err error) error {
	var strRet string

	if f == nil {
		return err
	}
	if f.IsDir() {
		return nil
	}

	strRet = path //+ "\r\n"

	//用strings.HasSuffix(src, suffix)//判断src中是否包含 suffix结尾
	ok := strings.HasSuffix(strRet, ".tpl")
	if ok {

		listfile = append(listfile, strRet) //将目录push到listfile []string中
	}

	return nil
}

func getFileList(path string) string {
	//var strRet string
	err := filepath.Walk(path, Listfunc) //

	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}

	return " "
}

func handler(indir string, outdir string, config string) error {
	var helm HelmChart

	data, err := ioutil.ReadFile(config)
	if err != nil {
		log.Println("Open config file err: ", err)
		return err
	}

	err = json.Unmarshal(data, &helm)
	if err != nil {
		log.Println("json Unmarshal err: ", err)
		return err
	}

	funcMap := template.FuncMap{
		"add": func(i int) int {
			return i + 1
		},
		"replace": func(i string) string {
			j := strings.Replace(i, "_", ".", -1)
			return j
		},
	}
	getFileList(indir)

	t := template.New("").Funcs(funcMap)
	fmt.Println(listfile)
	t, _ = t.ParseFiles(listfile...)
	for index, value := range listfile {

		path := strings.TrimSuffix(value, ".tpl")
		path = outdir + path
		fmt.Println("Index = ", index, ";Value = ", path)
		os.MkdirAll(filepath.Dir(path), 0755)
		f, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
		fmt.Println(filepath.Base(value))
		err = t.ExecuteTemplate(f, filepath.Base(value), helm)
		if err != nil {
			log.Println("executing template:", err)
			return err
		}

	}
	return nil

}

func main() {
	var intdir, outdir, config string
	flag.StringVar(&intdir, "i", "", "The directory of murano template")
	flag.StringVar(&outdir, "o", "", "The directory of murano package ")
	flag.StringVar(&config, "c", "", "The config file of helm ")

	flag.Parse()
	err := handler(intdir, outdir, config)
	if err != nil {
		log.Println("handler murano package fail:", err)
	}
}
