package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func getCmdLineArg() {
	var defaultForks int
	if cfg.Forks == 0 {
		defaultForks = 100
	} else {
		defaultForks = cfg.Forks
	}

	forks := flag.Int("f", defaultForks, "set concurrent num")
	destIPList := flag.String("d", "", "set target iplist file")

	interval := flag.Int("i", 0, "set wait interval after every batch running,maybe you will use it with -f,unit is second")

	debug := flag.Bool("v", false, "open debug mode")
	version := flag.Bool("V", false, "show current version")
	notColorPrint := flag.Bool("nc", false, "close color print")
	notBackOnCopy := flag.Bool("nb", false, "close backup when copy")
	authMethod := flag.String("m", "", "ssh connect auth method [password|sshkey|smart],default is smart")
	timeOut := flag.Int("t", int(cfg.TimeOut.Seconds()), "set ssh connect time out, unit is second")
	become := flag.Bool("b", false, "if run cmd as root")
	remoteRun := flag.Bool("r", false, "copy script file to remote and run")
	noNewline := flag.Bool("n", false, "print result without new line between ip and result")
	copy := flag.Bool("c", false, "only copy local file to remote machine's some directory[can config]")
	server = flag.Bool("server", false, "open server mode [not supported now]")
	client = flag.Bool("client", false, "open client mod [not supported now]")
	cronAdd := flag.Bool("cronadd", false, "add crontab job in current user ")
	cronDel := flag.Bool("crondel", false, "del crontab job in current user ")
	annotation := flag.String("a", "", "annotate the crontab task,useful with -cronadd")
	adminWeb := flag.Bool("admin", false, "open http admin web ")

	flag.Parse()
	other := flag.Args()
	timeOutInt := *timeOut
	if len(other) != 0 {
		cmd = other[0]
	}
	cfg.DestIPList = *destIPList
	cfg.AdminWeb = *adminWeb
	cfg.CronAdd = *cronAdd
	cfg.CronDel = *cronDel
	cfg.CronAnnotation = *annotation
	cfg.Forks = *forks
	cfg.Interval = *interval
	cfg.TimeOut = time.Second * time.Duration(timeOutInt)
	if *version {
		fmt.Println(cfg.Version)
		os.Exit(0)
	}
	if *remoteRun != false {
		cfg.RemoteRun = *remoteRun
	}
	if *noNewline != false {
		cfg.AddNewline = false
	}
	if *copy != false {
		cfg.Copy = *copy
	}
	if *debug != false {
		cfg.Debug = *debug
	}
	if *become != false {
		cfg.Become = *become
	}
	if *notColorPrint != false {
		cfg.ColorPrint = false
	}
	if *notBackOnCopy != false {
		cfg.BackOnCopy = false
	}
	if *authMethod != "" {
		cfg.AuthMethod = *authMethod
	}
	if cfg.Debug {
		fmt.Printf("[debug]%+v\n", cfg)
	}
}
