package log

import (
	"log"

	"github.com/fatih/color"
)

type level int

const (
	LevelDebug level = 0
	LevelInfo  level = 1
	LevelWarn  level = 2
	LevelError level = 3
	LevelFatal level = 4
)

var (
	currLevel level = LevelInfo

	prefixDegbug = color.CyanString("[DBG] ")
	prefixInfo   = color.GreenString("[INF] ")
	prefixWarn   = color.YellowString("[WAR] ")
	prefixError  = color.RedString("[ERR] ")
	prefixFatal  = color.MagentaString("[FAT] ")
	prefixPanic  = color.MagentaString("[PAN] ")
)

func SetLevel(l level) {
	currLevel = l
}

func Debug(msg string) {
	if LevelDebug >= currLevel {
		log.Println(prefixDegbug + msg)
	}
}

func Debugf(msg string, args ...interface{}) {
	if LevelDebug >= currLevel {
		log.Printf(prefixDegbug+msg, args...)
	}
}

func Info(msg string) {
	if LevelInfo >= currLevel {
		log.Println(prefixInfo + msg)
	}
}

func Infof(msg string, args ...interface{}) {
	if LevelInfo >= currLevel {
		log.Printf(prefixInfo+msg, args...)
	}
}

func Warn(msg string) {
	if LevelWarn >= currLevel {
		log.Println(prefixWarn + msg)
	}
}

func Warnf(msg string, args ...interface{}) {
	if LevelWarn >= currLevel {
		log.Printf(prefixWarn+msg, args...)
	}
}

func Error(msg string) {
	if LevelError >= currLevel {
		log.Println(prefixError + msg)
	}
}

func Errorf(msg string, args ...interface{}) {
	if LevelError >= currLevel {
		log.Printf(prefixError+msg, args...)
	}
}

func Fatal(msg string) {
	if LevelFatal >= currLevel {
		log.Fatalln(prefixFatal + msg)
	}
}

func Fatalf(msg string, args ...interface{}) {
	if LevelFatal >= currLevel {
		log.Fatalf(prefixFatal+msg, args...)
	}
}

/// Panic functions can't be disabled by setting the log level to a higher value
/// since they are meant to be used in critical situations where the application can't continue

func Panic(msg string) {
	log.Panicln(prefixPanic + msg)
}

func Panicf(msg string, args ...interface{}) {
	log.Panicf(prefixPanic+msg, args...)
}
