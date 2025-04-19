package utils

import "log"

func DebugLog(v ...any) {
	log.Println(" [debug] \t", v)
}

func ErrorLog(v ...any) {
	log.Println(" [error] \t", v)
}

func TraceLog(v ...any) {
	log.Fatal(" [trace] \t", v)
}
