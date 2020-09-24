package callposter

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func doConfFile(dir string) error {
	filename := fmt.Sprintf("%s/callposter.conf.xml", dir)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	xml.Unmarshal(content, &config)
	log.Println("config file loaded.")
	log.Println("http ip:", config.HTTP.IP)
	log.Println("http port:", config.HTTP.Port)
	log.Println("attribution Location:", config.Attribution.Location)
	log.Println("attribution hit prefix:", config.Attribution.HitPreifx)
	log.Println("attribution miss prefix:", config.Attribution.MissPrefix)
	return nil
}

func doDataFile(dir string) error {
	callerFilename := fmt.Sprintf("%s/外呼主叫号码表.xlsx", dir)
	//加载外呼主叫号码表
	f, err := excelize.OpenFile(callerFilename)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	rows, err := f.GetRows("外呼主叫号码表")
	if err != nil {
		log.Fatalln("外呼主叫号码表", err)
		return err
	}
	for _, row := range rows {
		for n, colCell := range row {
			if n == 0 && colCell != "" {
				// log.Println(n, ":", colCell)
				mng.mapping[colCell] = false
				callerDisplaySum++
			}
		}
	}
	log.Println("Caller display number excel file load success！ sum:", callerDisplaySum)

	//加载归属地二进制数据
	phoneFilename := fmt.Sprintf("%s/phone.dat", dir)
	content, err = ioutil.ReadFile(phoneFilename)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	totalLen = int32(len(content))
	firstOffset = get4(content[intLength : intLength*2])

	log.Println("phone attributtion data load success！")
	return nil
}

func doLogFile(dir string) error {
	filename := fmt.Sprintf("%s/callposter.conf.xml", dir)

	outfile, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	// 利用io.MultiWriter()将多个Writer拼成一个Writer使用
	mw := io.MultiWriter(os.Stdout, outfile)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile | log.LstdFlags)
	return nil
}

func doFile() error {
	flag.Parse()
	//配置文件
	err := doConfFile(*confDir)
	if err != nil {
		log.Println("do conf file error. exit...")
		return err
	}

	doLogFile(*logDir)
	if err != nil {
		log.Println("do log file error. exit...")
		return err
	}
	doDataFile(*dataDir)
	if err != nil {
		log.Println("do data file error. exit...")
		return err
	}
	return nil
}
