package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

func addCrontab(ip string) {
	if cmd != "" {
		copyAndRun(ip, rt.cronTmpFile)
	}
}

//对crontab定义的时间格式进行粗略的检测
func crontabFormatCheck(cmd string) (e error) {
	var b bool
	expectedFormatedString := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
		"11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23",
		"24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36",
		"37", "38", "39", "40", "41", "42", "43", "44", "45", "46", "47", "48", "49",
		"50", "51", "52", "53", "54", "55", "56", "57", "58", "59", "60", "-", "*", "/", ","}
	stringSlice := strings.Fields(cmd)
	//分隔后，长度至少6位
	if len(stringSlice) < 6 {
		e = errors.New("fileds is not enough")
		return
	}
	for _, i := range stringSlice[:5] {
		sequence := strings.Split(i, "")
		for _, s := range sequence {
			b = false
			for _, expected := range expectedFormatedString {
				if s == expected {
					b = true
				}
			}
			if b == false {
				e = errors.New(s + " is not expected string in " + cmd)
				return
			}
		}
	}
	return
}
func makeCronTmpFile(cmd string) (file string, e error) {
	var f *os.File
	var content string
	var annotation string
	var debug string
	annotation = setAnnotation(cfg.CronAnnotation)
	n := time.Now()
	suffix := n.Format("20060102150405")
	file = ".crontab." + suffix

	if f, e = os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0755); e != nil {
		return
	}
	defer f.Close()
	if cfg.Debug {
		debug = "set -x"
	} else {
		debug = ""
	}
	content = fmt.Sprintf(`#!/bin/bash
%s
user=$(id -u -n)
suffix=$(date +%%Y%%m%%d%%H%%M%%S)
pre="$(crontab -l 2>/dev/null)"
echo $pre > .cronbak.$suffix

rst=$(echo $pre|fgrep "%s")
#是否完全匹配
if [ "X$rst" != "X%s" ];then
  crontab - <<EOF
$pre
%s
%s
EOF

fi
`, debug, cmd, cmd, annotation, cmd)
	if _, e = f.WriteString(content); e != nil {
		return
	}
	return
}

func setAnnotation(annotation string) (comment string) {
	n := time.Now()
	date := n.Format("20060102150405")
	comment = "#Grun create:" + date + " " + annotation
	return
}
