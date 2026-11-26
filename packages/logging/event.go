package logging

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// ค่าของ LogConfiguration.Format
const (
	FormatNested = "nested" // ค่าเริ่มต้น (หรือค่าว่าง) แบบเดิม
	FormatLogfmt = "logfmt" // key=value บรรทัดเดียว
	FormatJSON   = "json"
)

// ชื่อ key มาตรฐานที่ Event() ใช้ ให้ทุก service ใช้ชื่อเดียวกันเพื่อค้นได้ตรงๆ
const (
	KeyEvent      = "event"
	KeyRequestId  = "request_id"
	KeyParentId   = "parent_id"
	KeyUserId     = "user_id"
	KeyBranchId   = "branch_id"
	KeyOrderId    = "order_id"
	KeyTarget     = "target"
	KeyMethod     = "method"
	KeyUrl        = "url"
	KeyPath       = "path"
	KeyStatus     = "status"
	KeyDurationMs = "duration_ms"
	KeyRetry      = "retry"
	KeyError      = "error"
	KeyCaller     = "caller"
)

// F คือ field ของ Event ในรูป key → value
type F map[string]any

// keyOrder ลำดับ key ที่ขึ้นก่อนใน output (ที่เหลือเรียงตามตัวอักษร) ให้คนอ่านเจอ event กับ id ก่อนเสมอ
var keyOrder = []string{
	logrus.FieldKeyTime, logrus.FieldKeyLevel, logrus.FieldKeyMsg,
	KeyEvent, KeyRequestId, KeyParentId, KeyUserId, KeyBranchId, KeyOrderId,
	KeyTarget, KeyMethod, KeyUrl, KeyPath, KeyStatus, KeyDurationMs, KeyRetry, KeyError,
}

var keyRank = func() map[string]int {
	m := make(map[string]int, len(keyOrder))
	for i, k := range keyOrder {
		m[k] = i
	}
	return m
}()

func sortKeys(keys []string) {
	sort.SliceStable(keys, func(i, j int) bool {
		ri, iok := keyRank[keys[i]]
		rj, jok := keyRank[keys[j]]
		switch {
		case iok && jok:
			return ri < rj
		case iok:
			return true
		case jok:
			return false
		}
		return keys[i] < keys[j]
	})
}

func newLogfmtFormatter() logrus.Formatter {
	return &logrus.TextFormatter{
		DisableColors:    true,
		FullTimestamp:    true,
		TimestampFormat:  "2006-01-02T15:04:05.000Z07:00",
		QuoteEmptyFields: true,
		SortingFunc:      sortKeys,
	}
}

// LevelForStatus คืน level ตาม HTTP status
// 0 (ต่อไม่ติด / timeout) และ 5xx = Error, 4xx = Warning, ที่เหลือ = Info
func LevelForStatus(status int) string {
	switch {
	case status == 0 || status >= 500:
		return Error
	case status >= 400:
		return Warning
	}
	return Info
}

// Event เขียน log 1 บรรทัดต่อ 1 เหตุการณ์ ชื่อ event เป็นค่าคงที่ (เช่น "outbound_call", "inbound_request")
// request_id, parent_id, user_id, branch_id, order_id ดึงจาก ctx ให้อัตโนมัติ (ค่าใน fields ทับค่าจาก ctx ได้)
//
// format nested (ค่าเริ่มต้น): ข้อความของบรรทัดเป็น key=value ทั้งชุด จึงค้นด้วย "request_id=..." ได้แม้ไม่ได้เปลี่ยน format
// format logfmt / json: ทุกค่าเป็น field จริงของ logrus
//
// ถ้าเปิด OnServerLog ไว้ จะส่งเข้า server log แบบไม่บล็อก พร้อม requestId / parentId
//
//go:noinline
func Event(ctx context.Context, logLevel string, event string, fields F) {
	_, file, line, functionName := getCaller()
	packageName, actualFunctionName := splitFuncName(functionName)

	data := eventFields(ctx, event, fields)

	var entry *logrus.Entry
	message := ""
	switch logconfig.Format {
	case FormatLogfmt, FormatJSON:
		data[KeyCaller] = fmt.Sprintf("%s:%d", shortFile(file), line)
		entry = logrus.WithFields(logrus.Fields{"application": logconfig.AppName})
		entry = entry.WithFields(logrus.Fields(data))
	default:
		entry = logrus.WithFields(logrus.Fields{
			"application": logconfig.AppName,
			"module":      packageName,
			"method":      actualFunctionName,
		})
		message = encodeLogfmt(data)
	}

	switch logLevel {
	case Debug:
		entry.Debug(message)
	case Warning:
		entry.Warn(message)
	case Error:
		entry.Error(message)
	case Fatal:
		// Event ไม่ปิดโปรแกรม: Fatal ถูกลดเป็น Error
		entry.Error(message)
	default:
		entry.Info(message)
	}

	if logconfig.OnServerLog && logLevel != Debug {
		trace := TraceFromContext(ctx)
		payload := make(map[string]any, len(data))
		for k, v := range data {
			payload[k] = v
		}
		go verifyLogger(logLevel, payload, trace.RequestId, trace.ParentId)
	}
}

// eventFields รวม field จาก ctx กับ fields ที่ส่งมา ค่า error ถูกแปลงเป็นข้อความเต็ม (ไม่ตัด)
func eventFields(ctx context.Context, event string, fields F) map[string]any {
	data := make(map[string]any, len(fields)+6)
	trace := TraceFromContext(ctx)
	setIfNotEmpty(data, KeyRequestId, trace.RequestId)
	setIfNotEmpty(data, KeyParentId, trace.ParentId)
	setIfNotEmpty(data, KeyUserId, trace.UserId)
	setIfNotEmpty(data, KeyBranchId, trace.BranchId)
	setIfNotEmpty(data, KeyOrderId, trace.OrderId)
	for k, v := range fields {
		if err, ok := v.(error); ok {
			if err == nil {
				continue
			}
			v = err.Error()
		}
		data[k] = v
	}
	data[KeyEvent] = event
	return data
}

func setIfNotEmpty(data map[string]any, key, value string) {
	if value != "" {
		data[key] = value
	}
}

// encodeLogfmt แปลง field เป็น "k=v k2=v2" เรียงตาม keyOrder ค่าที่มีช่องว่าง เครื่องหมาย = หรือ " จะถูก quote
func encodeLogfmt(data map[string]any) string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sortKeys(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(logfmtValue(data[k]))
	}
	return b.String()
}

func logfmtValue(v any) string {
	var s string
	switch t := v.(type) {
	case nil:
		return `""`
	case string:
		s = t
	case error:
		s = t.Error()
	default:
		s = fmt.Sprint(t)
	}
	if s == "" || strings.ContainsAny(s, " =\"\t\r\n") {
		return strconv.Quote(s)
	}
	return s
}

func splitFuncName(functionName string) (string, string) {
	lastDot := strings.LastIndex(functionName, ".")
	if lastDot == -1 {
		return "", functionName
	}
	return functionName[:lastDot], functionName[lastDot+1:]
}

// shortFile ตัด path เหลือ "dir/file.go" ให้บรรทัดไม่ยาวเกิน
func shortFile(file string) string {
	file = strings.ReplaceAll(file, "\\", "/")
	parts := strings.Split(file, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	return file
}
