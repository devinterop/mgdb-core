package logging

import (
	"bytes"
	"context"
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

// ctxKey เป็น type เฉพาะของ package นี้กันชนกับ context key ของ package อื่น (go vet เตือนถ้าใช้ string ดิบ)
type ctxKey string

// traceKey เป็น context key เดียวที่เก็บ TraceContext ทั้งก้อน — เพิ่ม field ตามรอย request ในอนาคต (traceId, spanId, ...)
// ได้โดยไม่ต้องเพิ่ม context key ใหม่
const traceKey ctxKey = "trace"

// TraceContext รวม field ที่ใช้ตามรอย request ไว้ในที่เดียว ให้ LoggerV3 อ่านทีเดียวจบ
type TraceContext struct {
	RequestId string
	ParentId  string
}

// WithTrace ผูก TraceContext ทั้งก้อนเข้ากับ ctx (ทับของเดิมทั้งหมดถ้ามี)
func WithTrace(ctx context.Context, trace TraceContext) context.Context {
	return context.WithValue(ctx, traceKey, trace)
}

// TraceFromContext อ่าน TraceContext ออกจาก ctx คืน zero value ถ้ายังไม่มีการ set (ctx เป็น nil ก็ปลอดภัย)
func TraceFromContext(ctx context.Context) TraceContext {
	if ctx == nil {
		return TraceContext{}
	}
	if trace, ok := ctx.Value(traceKey).(TraceContext); ok {
		return trace
	}
	return TraceContext{}
}

// WithRequestId ผูก requestId เข้ากับ ctx โดยคง field อื่นของ TraceContext เดิมไว้ (เช่น parentId ที่ set มาก่อนหน้า)
// ให้ LoggerV3 อ่านออกมา print ได้เองโดยไม่ต้องส่งซ้ำทุกครั้ง
// เรียกครั้งเดียวที่ middleware ต้นทางของ request (เช่น c.Request = c.Request.WithContext(logging.WithRequestId(c.Request.Context(), reqID)))
func WithRequestId(ctx context.Context, requestId string) context.Context {
	trace := TraceFromContext(ctx)
	trace.RequestId = requestId
	return WithTrace(ctx, trace)
}

// WithParentId ผูก parentId เข้ากับ ctx โดยคง field อื่นของ TraceContext เดิมไว้ (เช่น requestId ที่ set มาก่อนหน้า)
// เรียกที่ middleware ต้นทางของ request (เช่น กรณีอ่าน X-Parent-ID header มาได้)
func WithParentId(ctx context.Context, parentId string) context.Context {
	trace := TraceFromContext(ctx)
	trace.ParentId = parentId
	return WithTrace(ctx, trace)
}

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
		verifyLogger(logLevel, message, "", "")
	}
}

func verifyLogger(logLevel string, massage interface{}, requestId string, parentId string) {
	var logObj structs.JsonLogBody
	logObj.App_id = logconfig.AppId
	logObj.App_name = logconfig.AppName
	logObj.Level = getLogLevel(logLevel).String()
	logObj.Message = massage
	logObj.RequestId = requestId
	logObj.ParentId = parentId

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
		verifyLogger(logLevel, message, "", "")
	}
}

//go:noinline
func LoggerV3(ctx context.Context, logLevel string, message interface{}) {
	_, file, line, functionName := getCaller()

	lastDot := strings.LastIndex(functionName, ".")
	var packageName, actualFunctionName string
	if lastDot != -1 {
		packageName = functionName[:lastDot]
		actualFunctionName = functionName[lastDot+1:]
	} else {
		actualFunctionName = functionName
	}

	saveLog := false

	trace := TraceFromContext(ctx)
	requestId := trace.RequestId
	parentId := trace.ParentId

	logrusFields := logrus.Fields{
		"application": logconfig.AppName,
		"module":      packageName,
		"method":      actualFunctionName,
		"file":        file,
		"line":        line,
	}
	if requestId != "" {
		logrusFields["requestId"] = requestId
	}
	if parentId != "" {
		logrusFields["parentId"] = parentId
	}
	logEntry := logrus.WithFields(logrusFields)

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
		verifyLogger(logLevel, message, requestId, parentId)
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
