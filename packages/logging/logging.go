package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/devinterop/mgdb-core/app/structs"
	cnst "github.com/devinterop/mgdb-core/cnst"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

type LoggingServiceBackend struct{}

var Debug = cnst.Debug
var Info = cnst.Info
var Warning = cnst.Warning
var Error = cnst.Error
var Fatal = cnst.Fatal
var logconfig structs.LogConfiguration

// |2021-11-02T10:08:45.919+0700|INFO|documentapi|FormatControllers|CreateFormat|send mq api: http://xxxxxxxxx |
// 2021-11-02T10:08:45.919+0700 [INFO] [documentapi] [FormatControllers] [CreateFormat] send mq api: http://xxxxxxxxx

// https://stackoverflow.com/questions/48971780/how-to-change-the-format-of-log-output-in-logrus
func InitLog(logconfigInit structs.LogConfiguration) {
	logconfig = logconfigInit
	logLevel := logconfig.Level
	logrus.SetLevel(getLogLevel(logLevel))
	logrus.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		FieldsOrder:     []string{"application", "module", "method"},
		TimestampFormat: "2006-01-02T15:04:05.000",
		ShowFullLevel:   true,
		NoFieldsSpace:   false,
		NoFieldsColors:  false,
		NoColors:        false,
		TrimMessages:    false,
	})
}

func sendLogger(data structs.JsonLogBody) {
	defer func() {
		if err := recover(); err != nil {
			logrus.WithFields(logrus.Fields{
				"application": logconfig.AppName,
				"module":      "LoggingServiceBackend",
				"method":      "sendLogger",
			}).Error("panic occurred:", err)
		}
	}()
	sendToServerLog(data)
}

func sendToServerLog(data structs.JsonLogBody) {
	server := logconfig.Server
	servicePath := logconfig.ServicePath
	url := server + servicePath

	byteData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byteData))
	if err != nil {
		panic(err)
	}
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", os.Getenv("token"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result structs.JsonLogResponseEror
	json.Unmarshal([]byte(body), &result)
	logrus.WithFields(logrus.Fields{
		"application": logconfig.AppName,
		"module":      "LoggingServiceBackend",
		"method":      "sendToServerLog",
	}).Debug(fmt.Sprintf("result: %s, %s", resp.Status, body))
}

func Logger(logLevel string, message any, fields structs.LogrusField, saveLogOption ...bool) {
	saveLog := true
	if len(saveLogOption) > 0 {
		saveLog = saveLogOption[0]
	}

	fields.Application = logconfig.AppName

	logrusField := logrus.Fields{
		"application": fields.Application,
		"module":      fields.Module,
		"method":      fields.Method,
	}

	if fields.Module == "" {
		logrusField["module"] = "LoggingServiceBackend"
	}
	if fields.Method == "" {
		logrusField["method"] = "Logger"
	}

	entry := logrus.WithFields(logrusField)

	switch logLevel {
	case Debug:
		entry.Debug(message)
	case Info:
		entry.Info(message)
	case Warning:
		entry.Warn(message)
	case Error:
		entry.Error(message)
	case Fatal:
		entry.Fatal(message)
	default:
		entry.Info(message)
	}

	if logconfig.OnServerLog && saveLog {
		verifyLogger(logLevel, message)
	}
}

func verifyLogger(logLevel string, massage interface{}) {
	var logObj structs.JsonLogBody
	logObj.App_id = logconfig.AppId
	logObj.App_name = logconfig.AppName
	logObj.Level = getLogLevel(logLevel).String()
	logObj.Message = massage

	isServerLog := logconfig.OnServerLog
	logLevelEnv := logconfig.Level
	if isServerLog && logLevel != Debug {
		state := getLogState(logLevelEnv)
		if containInSlice(state, logLevel) {
			sendLogger(logObj)
		}
	}
}

func getLogState(level string) []string {
	debugLvl := []string{Debug, Info, Warning, Error, Fatal}
	infoLvl := []string{Info, Warning, Error, Fatal}
	warningLvl := []string{Warning, Error, Fatal}
	errorLvl := []string{Error, Fatal}
	fatalLvl := []string{Fatal}
	switch level {
	case Debug:
		return debugLvl
	case Info:
		return infoLvl
	case Warning:
		return warningLvl
	case Error:
		return errorLvl
	case Fatal:
		return fatalLvl
	}
	return nil
}

func getLogLevel(logLevel string) logrus.Level {
	switch logLevel {
	case Debug:
		return logrus.DebugLevel
	case Info:
		return logrus.InfoLevel
	case Warning:
		return logrus.WarnLevel
	case Error:
		return logrus.ErrorLevel
	case Fatal:
		return logrus.FatalLevel
	}
	return logrus.DebugLevel
}

func containInSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

//go:noinline
func LoggerV2(logLevel string, message interface{}, saveLogOption ...bool) {
	_, file, line, functionName := getCaller()

	lastDot := strings.LastIndex(functionName, ".")
	var packageName, actualFunctionName string
	if lastDot != -1 {
		packageName = functionName[:lastDot]
		actualFunctionName = functionName[lastDot+1:]
	} else {
		actualFunctionName = functionName
	}

	saveLog := true
	if len(saveLogOption) > 0 {
		saveLog = saveLogOption[0]
	}

	logEntry := logrus.WithFields(logrus.Fields{
		"application": logconfig.AppName,
		"module":      packageName,
		"method":      actualFunctionName,
		"file":        file,
		"line":        line,
	})

	switch logLevel {
	case Debug:
		logEntry.Debug(message)
	case Info:
		logEntry.Info(message)
	case Warning:
		logEntry.Warn(message)
	case Error:
		logEntry.Error(message)
	case Fatal:
		logEntry.Fatal(message)
	default:
		logEntry.Info(message)
	}

	if logconfig.OnServerLog && saveLog {
		verifyLogger(logLevel, message)
	}
}

// getCaller walks up the call stack skipping frames within this logging package.
func getCaller() (uintptr, string, int, string) {
	const loggingPkg = "github.com/devinterop/mgdb-core/packages/logging"
	for i := 1; i < 10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		name := fn.Name()
		if !strings.Contains(name, loggingPkg) {
			return pc, file, line, name
		}
	}
	return 0, "?", 0, "?"
}
