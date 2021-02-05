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

func delCrontab(ip string) {
	if cmd != "" {
		copyAndRun(ip, rt.cronTmpFile)
	}
}

//在添加crontab时，对crontab定义的时间格式和命令格式进行粗略的检测
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
	//前5位应当是expectedFormatedString中定义的字符
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
	//禁止命令是rm -rf /,rm -rf /*，rm /,rm /*
	if stringSlice[5] == "rm" {
		if stringSlice[6] == "-rf" && stringSlice[7] == "/" {
			e = errors.New(stringSlice[5] + " " + stringSlice[6] + " " + stringSlice[7] + " is not allow in crontab")
			return
		}
		if stringSlice[6] == "-rf" && stringSlice[7] == "/*" {
			e = errors.New(stringSlice[5] + " " + stringSlice[6] + " " + stringSlice[7] + " is not allow in crontab")
			return
		}
		if stringSlice[6] == "/" {
			e = errors.New(stringSlice[5] + " " + stringSlice[6] + " is not allow in crontab")
			return
		}
		if stringSlice[6] == "/*" {
			e = errors.New(stringSlice[5] + " " + stringSlice[6] + " is not allow in crontab")
			return
		}
	}
	return

}

//在删除crontab的时候，对crontab定义的时间格式进行粗略的检测，不检测是否包含rm /
func crontabFormatCheckForDel(cmd string) (e error) {
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
	//前5位应当是expectedFormatedString中定义的字符
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
func convertCronCmd(cmd string) (newcmd string) {
	cmdSlice := strings.Fields(cmd)
	newcmd = strings.Join(cmdSlice, " ")
	return
}
func makeCronAddTmpFile(cmd string) (file string, e error) {
	var f *os.File
	var content string
	var annotation string
	var debug string
	var tail string
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
		tail = "crontab -l"

	} else {
		debug = ""
		tail = ""
	}
	content = fmt.Sprintf(`#!/bin/bash
%s
user=$(id -u -n)
suffix=$(date +%%Y%%m%%d%%H%%M%%S)
pre="$(crontab -l 2>/dev/null)"
echo "$pre" > .cronbak.$suffix
#检测新增任务是否已经存在
rst=$(echo "$pre"|fgrep "%s"|head -n1)
#存在的任务是否完全匹配新增任务，只有完全匹配才不新增
if [ "X$rst" != "X%s" ];then
  crontab - <<EOF
$pre
%s
%s
EOF
else
  echo "cron job  already exists!"
fi

%s
`, debug, cmd, cmd, annotation, cmd, tail)
	if _, e = f.WriteString(content); e != nil {
		return
	}
	return
}

func makeCronDelTmpFile(cmd string) (file string, e error) {
	var f *os.File
	var content string
	var debug string
	var tail string
	n := time.Now()
	suffix := n.Format("20060102150405")
	file = ".crontab." + suffix

	if f, e = os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0755); e != nil {
		return
	}
	defer f.Close()
	if cfg.Debug {
		debug = "set -x"
		tail = "crontab -l"
	} else {
		debug = ""
		tail = ""
	}
	content = fmt.Sprintf(`#!/bin/bash
%s
#备份crontab -l
suffix=$(date +%%Y%%m%%d%%H%%M%%S)
prelist="$(crontab -l 2>/dev/null)"
echo ""
echo "$prelist" > .cronbak.$suffix

rmline="%s"
pre=""
newText=""
#循环拿当前行和上一行
while read line;do
  pre="${current}"
  current="$line"
  if [ "x$current" = "x$rmline" ] ;then
    if [[ "${pre}" =~ ^"#Grun create:" ]];then
      pre=""
    fi
    current=""
  fi

  if [ "${pre}" != "" ];then
    newText="${newText}""${pre}\n"
  fi

done <<<"$prelist"
#循环后把最后一行再追加上
if [ "${current}" != "" ];then
  newText="${newText}""${current}\n"
fi
#重新导入crontab
echo -ne "$newText" |crontab -
%s
`, debug, cmd, tail)
	if cfg.Debug {
		fmt.Println("tmp file:", file)
	}
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
