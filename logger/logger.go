package logger

import "log"

var loglib logger

var Verbose bool

type logger interface {
	Infoln(args ...interface{})
	Errorln(args ...interface{})
}

// 外部からロガーを設定
func SetLogger(l logger) {
	loglib = l
}

// デバッグ info errorの三段階のログレベルに対応

func Debugln(args ...interface{}) {
	if Verbose {
		log.Println(args...)
	}
}

func Infoln(args ...interface{}) {
	if loglib != nil {
		loglib.Infoln(args...)
	} else if Verbose {
		log.Println(args...)
	}
}

func Errorln(args ...interface{}) {
	if loglib != nil {
		loglib.Errorln(args...)
	} else if Verbose {
		log.Println(args...)
	}
}
