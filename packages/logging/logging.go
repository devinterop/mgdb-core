package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"github.com/devinterop/mgdb-core/app/structs"
	cnst "github.com/devinterop/mgdb-core/cnst"

	//"github.com/devinterop/mgdb-core/utils"
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
	// logrus.SetFormatter(&logrus.JSONFormatter{})
	// logrus.SetFormatter(&logrus.TextFormatter{
	// 	FullTimestamp: true,
	// 	ForceColors:   true,
	// 	DisableColors: false,
	// })
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
	servicePath := logconfig.ServicePath //"api/logging"
	url := server + servicePath

	byteData, err := json.Marshal(data)
	if err != nil {
		//logrus.Error("err: ", err)
		panic(err)
	}
	//var jsonStr = []byte(string(byteData))
	// log.Println("send log json :", string(byteData))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byteData))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", os.Getenv("token"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result structs.JsonLogResponseEror
	json.Unmarshal([]byte(body), &result)
	//fmt.Printf("%+v\n", result)
	logrus.WithFields(logrus.Fields{
		"application": logconfig.AppName,
		"module":      "LoggingServiceBackend",
		"method":      "sendToServerLog",
	}).Debug(fmt.Sprintf("result: %s, %s", resp.Status, body))
}

func Logger(logLevel string, massage interface{}, fields structs.LogrusField, saveLog_option ...bool) {
	saveLog := true
	if len(saveLog_option) > 0 {
		saveLog = saveLog_option[0]
	}
	isServerLog := logconfig.OnServerLog

	fields.Application = logconfig.AppName
	fieldsJSON, err := json.Marshal(fields)
	if err != nil {
		//panic(err)
		logrus.WithFields(logrus.Fields{
			"application": logconfig.AppName,
			"module":      "LoggingServiceBackend",
			"method":      "Logger",
		}).Error("panic occurred: ", err)
	}
	jsonMap := make(map[string]interface{})
	err2 := json.Unmarshal(fieldsJSON, &jsonMap)
	if err2 != nil {
		//panic(err)
		logrus.WithFields(logrus.Fields{
			"application": logconfig.AppName,
			"module":      "LoggingServiceBackend",
			"method":      "Logger",
		}).Error("panic occurred: ", err)
	}

	// set default logrusField if fields conversion has error
	var logrusField logrus.Fields
	if err != nil || err2 != nil {
		logrusField = logrus.Fields{
			"application": logconfig.AppName,
		}
		jsonMap = logrusField
	}

	if logLevel == Debug {
		logrus.WithFields(jsonMap).Debug(massage)
		if isServerLog && saveLog {
			verifyLogger(logLevel, massage)
		}
	} else if logLevel == Info {
		logrus.WithFields(jsonMap).Info(massage)
		if isServerLog && saveLog {
			verifyLogger(logLevel, massage)
		}
	} else if logLevel == Warning {
		logrus.WithFields(jsonMap).Warn(massage)
		if isServerLog && saveLog {
			verifyLogger(logLevel, massage)
		}
	} else if logLevel == Error {
		logrus.WithFields(jsonMap).Error(massage)
		if isServerLog && saveLog {
			verifyLogger(logLevel, massage)
		}
	} else if logLevel == Fatal {
		logrus.WithFields(jsonMap).Fatal(massage)
		if isServerLog && saveLog {
			verifyLogger(logLevel, massage)
		}
	}
}

func LoggerV2(logLevel string, message interface{}, saveLogOption ...bool) {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return
	}
	function := runtime.FuncForPC(pc)
	functionName := function.Name()

	// Extract the base name of the file
	// fileNameWithExt := filepath.Base(file)
	// // Remove the file extension
	// fileName := strings.TrimSuffix(fileNameWithExt, filepath.Ext(fileNameWithExt))

	// Get the struct name using reflection
	// structType := reflect.TypeOf(structInstance)
	// structName := ""
	// if structType.Kind() == reflect.Ptr {
	// 	structName = structType.Elem().Name()
	// } else {
	// 	structName = structType.Name()
	// }

	saveLog := true
	if len(saveLogOption) > 0 {
		saveLog = saveLogOption[0]
	}
	isServerLog := logconfig.OnServerLog
	fields := structs.LogrusField{}
	fields.Application = logconfig.AppName
	// fields.Module = structName
	fields.Method = functionName
	fields.File = file
	fields.Line = line
	fieldsJSON, err := json.Marshal(fields)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"application": logconfig.AppName,
			"module":      "LoggingServiceBackend",
			"method":      "Logger",
		}).Error("Error marshaling fields: ", err)
		return
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(fieldsJSON, &jsonMap); err != nil {
		logrus.WithFields(logrus.Fields{
			"application": logconfig.AppName,
			"module":      "LoggingServiceBackend",
			"method":      "Logger",
		}).Error("Error unmarshaling fieldsJSON: ", err)
		jsonMap = map[string]interface{}{
			"application": logconfig.AppName,
		}
	}

	logEntry := logrus.WithFields(jsonMap)

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
		logEntry.Info(message) // Default to Info if unknown log level
	}

	if isServerLog && saveLog {
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
