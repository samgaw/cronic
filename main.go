package main

import (
	"flag"
	"fmt"
	"github.com/samgaw/cronic/cron"
	"github.com/samgaw/cronic/crontab"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] CRONTAB\n\nAvailable options:\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	debug := flag.Bool("debug", false, "enable debug logging")
	json := flag.Bool("json", false, "enable JSON logging")
	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if *json {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	}

	if flag.NArg() != 1 {
		Usage()
		os.Exit(2)
		return
	}

	crontabFileName := flag.Args()[0]
	logrus.Infof("read crontab: %s", crontabFileName)

	tab, err := readCrontabAtPath(crontabFileName)

	if err != nil {
		logrus.Fatal(err)
		return
	}

	var (
		wg        sync.WaitGroup
		exitChans []chan interface{}
	)

	for _, job := range tab.Jobs {
		exitChan := make(chan interface{}, 1)
		exitChans = append(exitChans, exitChan)

		cronLogger := logrus.WithFields(logrus.Fields{
			"job.schedule": job.Schedule,
			"job.command":  job.Command,
			"job.position": job.Position,
		})

		cron.StartJob(&wg, tab.Context, job, exitChan, cronLogger)
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	termSig := <-termChan

	logrus.Infof("received %s, shutting down", termSig)
	for _, c := range exitChans {
		c <- true
	}

	logrus.Info("waiting for jobs to finish")
	wg.Wait()

	logrus.Info("exiting")
}

func readCrontabAtPath(path string) (*crontab.Crontab, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	return crontab.ParseCrontab(file)
}
