package model_name_service

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/devinterop/mgdb-core/app/structs"
	"github.com/devinterop/mgdb-core/packages/logging"
	"github.com/sirupsen/logrus"
)

func captureLogfmt(t *testing.T, level string) *bytes.Buffer {
	t.Helper()
	oldOut := logrus.StandardLogger().Out
	oldFormatter := logrus.StandardLogger().Formatter
	oldLevel := logrus.GetLevel()
	oldBackoff := retryBackoff
	t.Cleanup(func() {
		logging.InitLog(structs.LogConfiguration{})
		logrus.SetOutput(oldOut)
		logrus.SetFormatter(oldFormatter)
		logrus.SetLevel(oldLevel)
		retryBackoff = oldBackoff
	})
	retryBackoff = time.Millisecond
	logging.InitLog(structs.LogConfiguration{AppName: "testapp", Level: level, Format: logging.FormatLogfmt})
	buf := &bytes.Buffer{}
	logrus.SetOutput(buf)
	return buf
}

func reqCtx() context.Context {
	ctx := logging.WithRequestId(context.Background(), "posapi-REQ1")
	return logging.WithOrderId(ctx, "order-5")
}

func lines(buf *bytes.Buffer) []string {
	s := strings.TrimSpace(buf.String())
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func TestSendApiHeadersCtx_OneLineAndForwardsRequestId(t *testing.T) {
	buf := captureLogfmt(t, logging.Debug)
	var gotReqId, gotParentId, gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReqId = r.Header.Get(HeaderRequestId)
		gotParentId = r.Header.Get(HeaderParentId)
		gotKey = r.Header.Get("API_KEY")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	body, status := ServiceModel{}.SendApiHeadersCtx(reqCtx(), "ax", []byte(`{}`), srv.URL+"/AX2012/GetOnhand?item=A01", http.MethodGet,
		map[string]interface{}{"API_KEY": "secret"})

	if body != `{"ok":true}` || status != "200 OK" {
		t.Fatalf("unexpected result: %q %q", body, status)
	}
	if gotReqId != "posapi-REQ1" || gotParentId != "posapi-REQ1" || gotKey != "secret" {
		t.Errorf("headers not forwarded: reqId=%q parentId=%q key=%q", gotReqId, gotParentId, gotKey)
	}
	ls := lines(buf)
	if len(ls) != 1 {
		t.Fatalf("want 1 log line, got %d: %q", len(ls), buf.String())
	}
	for _, want := range []string{"level=info", "event=outbound_call", "request_id=posapi-REQ1", "order_id=order-5",
		"target=ax", "method=GET", "/AX2012/GetOnhand", "status=200", "duration_ms=", "retry=0"} {
		if !strings.Contains(ls[0], want) {
			t.Errorf("missing %q in %q", want, ls[0])
		}
	}
	if strings.Contains(ls[0], "item=A01") || strings.Contains(ls[0], "secret") {
		t.Errorf("query string / header value must not be logged: %q", ls[0])
	}
}

func TestSendApiHeadersCtx_KeepsCallerRequestIdHeader(t *testing.T) {
	captureLogfmt(t, logging.Debug)
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get(HeaderRequestId)
	}))
	defer srv.Close()

	ServiceModel{}.SendApiHeadersCtx(reqCtx(), "phr", nil, srv.URL, http.MethodPost, map[string]interface{}{HeaderRequestId: "caller-set"})
	if got != "caller-set" {
		t.Errorf("caller header overwritten: %q", got)
	}
}

func TestSendApiHeadersCtx_4xxIsWarnEvenWithDebugOption(t *testing.T) {
	buf := captureLogfmt(t, logging.Info)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/missing") {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// สำเร็จ + WithLogLevel(Debug) ต้องไม่ออกที่ level info
	ServiceModel{}.SendApiHeadersCtx(reqCtx(), "ax", nil, srv.URL+"/ok", http.MethodGet, nil, WithLogLevel(logging.Debug))
	if buf.Len() != 0 {
		t.Fatalf("success with debug level should be hidden at info: %q", buf.String())
	}

	_, status := ServiceModel{}.SendApiHeadersCtx(reqCtx(), "ax", nil, srv.URL+"/missing", http.MethodGet, nil, WithLogLevel(logging.Debug))
	if status != "404 Not Found" {
		t.Errorf("status = %q", status)
	}
	if !strings.Contains(buf.String(), "level=warning") || !strings.Contains(buf.String(), "status=404") {
		t.Errorf("404 should be logged as warning: %q", buf.String())
	}
}

func TestSendApiHeadersCtx_5xxIsError(t *testing.T) {
	buf := captureLogfmt(t, logging.Info)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ServiceModel{}.SendApiHeadersCtx(reqCtx(), "personal", nil, srv.URL, http.MethodGet, nil)
	if !strings.Contains(buf.String(), "level=error") || !strings.Contains(buf.String(), "status=500") {
		t.Errorf("500 should be error: %q", buf.String())
	}
}

func TestSendApiHeadersCtx_Timeout(t *testing.T) {
	buf := captureLogfmt(t, logging.Info)
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)

	start := time.Now()
	body, status := ServiceModel{}.SendApiHeadersCtx(reqCtx(), "nodered", nil, srv.URL, http.MethodPost, nil, WithTimeout(50*time.Millisecond))
	if time.Since(start) > 2*time.Second {
		t.Fatal("timeout not applied")
	}
	if body != "" || status != "500" {
		t.Errorf("timeout should return same as legacy error: %q %q", body, status)
	}
	out := buf.String()
	for _, want := range []string{"level=error", "status=0", "target=nodered", "timeout after 50ms"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %q", want, out)
		}
	}
}

// connResetServer ปิด connection ทันทีโดยไม่ตอบ เพื่อจำลอง connection reset
func connResetServer(t *testing.T, hits *int32) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			atomic.AddInt32(hits, 1)
			c.Close()
		}
	}()
	return "http://" + ln.Addr().String()
}

func TestDoOutbound_RetryOnlyIdempotent(t *testing.T) {
	buf := captureLogfmt(t, logging.Info)

	var getHits int32
	getUrl := connResetServer(t, &getHits)
	res := ServiceModel{}.DoOutbound(reqCtx(), OutboundRequest{Target: "ax", Method: http.MethodGet, Url: getUrl}, WithRetry(2))
	if res.Err == nil || res.Retry != 2 || atomic.LoadInt32(&getHits) != 3 {
		t.Errorf("GET should retry 2 times: err=%v retry=%d hits=%d", res.Err, res.Retry, getHits)
	}
	if !strings.Contains(buf.String(), "retry=2") {
		t.Errorf("retry count not logged: %q", buf.String())
	}

	var postHits int32
	postUrl := connResetServer(t, &postHits)
	res = ServiceModel{}.DoOutbound(reqCtx(), OutboundRequest{Target: "nodered", Method: http.MethodPost, Url: postUrl}, WithRetry(2))
	if res.Retry != 0 || atomic.LoadInt32(&postHits) != 1 {
		t.Errorf("POST must not retry: retry=%d hits=%d", res.Retry, postHits)
	}
	if len(lines(buf)) != 2 {
		t.Errorf("want one line per call: %q", buf.String())
	}
}

func TestDoOutbound_NoRetryOn5xx(t *testing.T) {
	captureLogfmt(t, logging.Info)
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	res := ServiceModel{}.DoOutbound(reqCtx(), OutboundRequest{Target: "ax", Method: http.MethodGet, Url: srv.URL}, WithRetry(3))
	if res.StatusCode != http.StatusBadGateway || hits != 1 {
		t.Errorf("5xx must not retry: status=%d hits=%d", res.StatusCode, hits)
	}
}

func TestRedactUrl(t *testing.T) {
	cases := map[string]string{
		"http://u:p@host/a/b?phone=0812345678#x": "http://host/a/b",
		"http://host/path":                       "http://host/path",
		"::bad url?x=1":                          "::bad url",
	}
	for in, want := range cases {
		if got := redactUrl(in); got != want {
			t.Errorf("redactUrl(%q) = %q, want %q", in, got, want)
		}
	}
}
