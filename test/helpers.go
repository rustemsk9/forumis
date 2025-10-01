package test

import "log"

var logger *log.Logger

func Info(args ...interface{}) {
	logger.SetPrefix("[INFO] ")
	logger.Println(args...)
}

func Danger(args ...interface{}) {
	logger.SetPrefix("[ERROR] ")
	logger.Println(args...)
}

func Warn(args ...interface{}) {
	logger.SetPrefix("[WARN] ")
	logger.Println(args...)
}
