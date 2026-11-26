package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/devinterop/mgdb-core/app/structs"
	"github.com/sirupsen/logrus"
)

// captureLog ตั้ง config + format แล้วคืน buffer ที่รับ output ของ logrus
func captureLog(t *testing.T, format string, level string) *bytes.Buffer {
	t.Helper()
	oldConfig := logconfig
	oldOut := logrus.StandardLogger().Out
	oldFormatter := logrus.StandardLogger().Formatter
	oldLevel := logrus.GetLevel()
	t.Cleanup(func() {
		logconfig = oldConfig
		logrus.SetOutput(oldOut)
		logrus.SetFormatter(oldFormatter)
		logrus.SetLevel(oldLevel)
	})

	InitLog(structs.LogConfiguration{AppName: "testapp", Level: level, Format: format})
	buf := &bytes.Buffer{}
	logrus.SetOutput(buf)
	return buf
}

func traceCtx() context.Context {
	ctx := WithRequestId(context.Background(), "posapi-AB12CD34")
	ctx = WithParentId(ctx, "web-XYZ")
	ctx = WithUser(ctx, "user-1", "branch-9")
	return WithOrderId(ctx, "order-77")
}

func TestEvent_Logfmt_OneLineWithContextFields(t *testing.T) {
	buf := captureLog(t, FormatLogfmt, Debug)

	Event(traceCtx(), Warning, "outbound_call", F{
		KeyTarget: "ax", KeyStatus: 404, KeyDurationMs: int64(21), KeyUrl: "http://ax/AX2012/GetOnhand",
	})

	out := buf.String()
	if strings.Count(out, "\n") != 1 {
		t.Fatalf("want exactly one line, got %q", out)
	}
	for _, want := range []string{
		"level=warning", "event=outbound_call", "request_id=posapi-AB12CD34", "parent_id=web-XYZ",
		"user_id=user-1", "branch_id=branch-9", "order_id=order-77", "target=ax", "status=404",
		"duration_ms=21", "url=\"http://ax/AX2012/GetOnhand\"", "application=testapp", "caller=",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %q", want, out)
		}
	}
	// event และ request_id ต้องขึ้นก่อน field อื่น
	if strings.Index(out, "event=") > strings.Index(out, "request_id=") ||
		strings.Index(out, "request_id=") > strings.Index(out, "target=") {
		t.Errorf("unexpected key order: %q", out)
	}
	if strings.Contains(out, "msg=") {
		t.Errorf("empty msg should be omitted: %q", out)
	}
}

func TestEvent_Nested_MessageIsKeyValue(t *testing.T) {
	buf := captureLog(t, "", Debug)

	Event(traceCtx(), Info, "onhand_result", F{"qty": 12, "source": "cache"})

	out := buf.String()
	for _, want := range []string{"event=onhand_result", "request_id=posapi-AB12CD34", "qty=12", "source=cache", "[INFO]"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %q", want, out)
		}
	}
}

func TestEvent_JSON_FieldsAreReal(t *testing.T) {
	buf := captureLog(t, FormatJSON, Debug)

	Event(traceCtx(), Error, "close_order_list", F{KeyError: errors.New("json: cannot unmarshal number into Go struct field X.total of type string")})

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("not json: %v (%q)", err, buf.String())
	}
	if m[KeyEvent] != "close_order_list" || m[KeyRequestId] != "posapi-AB12CD34" || m["level"] != "error" {
		t.Errorf("unexpected fields: %v", m)
	}
	if !strings.Contains(m[KeyError].(string), "X.total of type string") {
		t.Errorf("error should be full text, got %v", m[KeyError])
	}
}

func TestEvent_FieldsOverrideContext_AndNoContext(t *testing.T) {
	buf := captureLog(t, FormatLogfmt, Debug)

	Event(traceCtx(), Info, "x", F{KeyOrderId: "order-override"})
	Event(nil, Info, "y", nil) //nolint:staticcheck // ctx nil ต้องไม่ panic

	out := buf.String()
	if !strings.Contains(out, "order_id=order-override") || strings.Contains(out, "order_id=order-77") {
		t.Errorf("explicit field should override ctx: %q", out)
	}
	if !strings.Contains(out, "event=y") {
		t.Errorf("nil ctx event missing: %q", out)
	}
}

func TestEvent_RespectsLevel(t *testing.T) {
	buf := captureLog(t, FormatLogfmt, Info)

	Event(traceCtx(), Debug, "noisy", nil)
	if buf.Len() != 0 {
		t.Errorf("debug event should be filtered at info level: %q", buf.String())
	}
}

func TestEvent_QuotesValues(t *testing.T) {
	buf := captureLog(t, "", Debug)

	Event(context.Background(), Info, "x", F{"msgtext": "has space", "empty": "", "eq": "a=b"})
	out := buf.String()
	for _, want := range []string{`msgtext="has space"`, `empty=""`, `eq="a=b"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %q", want, out)
		}
	}
}

func TestLevelForStatus(t *testing.T) {
	cases := map[int]string{0: Error, 200: Info, 302: Info, 400: Warning, 404: Warning, 499: Warning, 500: Error, 503: Error}
	for status, want := range cases {
		if got := LevelForStatus(status); got != want {
			t.Errorf("LevelForStatus(%d) = %s, want %s", status, got, want)
		}
	}
}

func TestInitLog_DefaultFormatUnchanged(t *testing.T) {
	captureLog(t, "", Info)
	if _, ok := logrus.StandardLogger().Formatter.(*logrus.TextFormatter); ok {
		t.Fatal("default format must stay nested formatter")
	}
}

func TestWithHelpers_KeepOtherFields(t *testing.T) {
	ctx := traceCtx()
	ctx = WithRequestId(ctx, "new-id")
	tr := TraceFromContext(ctx)
	if tr.RequestId != "new-id" || tr.ParentId != "web-XYZ" || tr.UserId != "user-1" || tr.BranchId != "branch-9" || tr.OrderId != "order-77" {
		t.Errorf("unexpected trace: %+v", tr)
	}
}
