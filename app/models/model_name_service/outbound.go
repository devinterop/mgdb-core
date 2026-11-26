package model_name_service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/devinterop/mgdb-core/packages/logging"
)

// DefaultOutboundTimeout timeout ต่อครั้งของการเรียกผ่าน SendApiHeadersCtx / DoOutbound ถ้าไม่ได้ส่ง WithTimeout
const DefaultOutboundTimeout = 30 * time.Second

// header ที่ส่งต่อ request id ให้ service ปลายทาง
const (
	HeaderRequestId = "X-Request-ID"
	HeaderParentId  = "X-Parent-ID"
)

// outboundClient ใช้ร่วมกันทั้ง package เพื่อ reuse connection
// ไม่ตั้ง Timeout ที่ client เพราะคุม timeout ต่อครั้งผ่าน context แทน
var outboundClient = &http.Client{}

// retryBackoff เวลารอก่อน retry ครั้งที่ n คือ n × retryBackoff
var retryBackoff = 200 * time.Millisecond

type outboundConfig struct {
	timeout      time.Duration
	retry        int
	successLevel string
}

// OutboundOption ตัวเลือกของ SendApiHeadersCtx / DoOutbound
type OutboundOption func(*outboundConfig)

// WithTimeout กำหนด timeout ต่อครั้ง (ค่า <= 0 ใช้ DefaultOutboundTimeout)
func WithTimeout(d time.Duration) OutboundOption {
	return func(c *outboundConfig) {
		if d > 0 {
			c.timeout = d
		}
	}
}

// WithRetry ให้ลองใหม่ได้สูงสุด n ครั้ง เฉพาะเมื่อต่อไม่ติด (network error / timeout)
// และเฉพาะ method ที่ทำซ้ำได้ปลอดภัย (GET, HEAD, OPTIONS) method อื่นจะไม่ retry เพื่อกันข้อมูลซ้ำที่ปลายทาง
func WithRetry(n int) OutboundOption {
	return func(c *outboundConfig) {
		if n > 0 {
			c.retry = n
		}
	}
}

// WithLogLevel กำหนด level ของ log เมื่อเรียกสำเร็จ (status < 400) เช่น Debug สำหรับการเรียกที่ถี่มาก
// ถ้าได้ 4xx / 5xx / ต่อไม่ติด จะใช้ logging.LevelForStatus เสมอ error จึงไม่ถูกซ่อน
func WithLogLevel(level string) OutboundOption {
	return func(c *outboundConfig) {
		c.successLevel = level
	}
}

// OutboundRequest ข้อมูลการเรียก service อื่น
type OutboundRequest struct {
	Target  string // ชื่อ service ปลายทาง เช่น phr, personal, nodered, ax
	Method  string
	Url     string
	Body    []byte
	Headers map[string]interface{}
}

// OutboundResult ผลการเรียก StatusCode = 0 เมื่อต่อไม่ติดหรือ timeout
type OutboundResult struct {
	Body       []byte
	StatusCode int
	Status     string // เช่น "200 OK"
	Duration   time.Duration
	Retry      int // จำนวนครั้งที่ลองใหม่
	Err        error
}

// DoOutbound เรียก service อื่นแล้ว log 1 บรรทัด event=outbound_call (target, method, url, status, duration_ms, retry, error)
// พร้อม request_id / order_id จาก ctx และส่ง X-Request-ID / X-Parent-ID ต่อให้ปลายทาง (ไม่ทับถ้า headers มีอยู่แล้ว)
func (s ServiceModel) DoOutbound(ctx context.Context, req OutboundRequest, opts ...OutboundOption) OutboundResult {
	if ctx == nil {
		ctx = context.Background()
	}
	cfg := outboundConfig{timeout: DefaultOutboundTimeout}
	for _, opt := range opts {
		opt(&cfg)
	}
	method := strings.ToUpper(req.Method)
	if method == "" {
		method = http.MethodGet
	}
	maxRetry := 0
	if isIdempotent(method) {
		maxRetry = cfg.retry
	}

	start := time.Now()
	var result OutboundResult
	for attempt := 0; ; attempt++ {
		result = doOnce(ctx, method, req, cfg.timeout)
		result.Retry = attempt
		if result.Err == nil || attempt >= maxRetry || ctx.Err() != nil || !isRetryable(result) {
			break
		}
		select {
		case <-time.After(time.Duration(attempt+1) * retryBackoff):
		case <-ctx.Done():
		}
		if ctx.Err() != nil {
			break
		}
	}
	result.Duration = time.Since(start)

	level := logging.LevelForStatus(result.StatusCode)
	if result.Err == nil && result.StatusCode < 400 && cfg.successLevel != "" {
		level = cfg.successLevel
	}
	fields := logging.F{
		logging.KeyTarget:     req.Target,
		logging.KeyMethod:     method,
		logging.KeyUrl:        redactUrl(req.Url),
		logging.KeyStatus:     result.StatusCode,
		logging.KeyDurationMs: result.Duration.Milliseconds(),
		logging.KeyRetry:      result.Retry,
	}
	if result.Err != nil {
		fields[logging.KeyError] = result.Err
	}
	logging.Event(ctx, level, "outbound_call", fields)
	return result
}

// SendApiHeadersCtx ใช้แทน SendApiHeaders ได้ทันที คืนค่ารูปแบบเดียวกัน (body, status เช่น "200 OK")
// และคืน "500" เมื่อต่อไม่ติดหรืออ่าน body ไม่ได้ เหมือนของเดิม
// ต่างจากของเดิมตรงที่ log 1 บรรทัดต่อการเรียก มี timeout และส่ง request id ต่อให้ปลายทาง
func (s ServiceModel) SendApiHeadersCtx(ctx context.Context, target string, data []byte, url string, requestType string, headers map[string]interface{}, opts ...OutboundOption) (string, string) {
	result := s.DoOutbound(ctx, OutboundRequest{
		Target:  target,
		Method:  requestType,
		Url:     url,
		Body:    data,
		Headers: headers,
	}, opts...)
	if result.Err != nil {
		// ต่อไม่ติด / timeout / อ่าน body ไม่ได้ ของเดิมคืน "500" ทุกกรณี
		return "", "500"
	}
	return string(result.Body), result.Status
}

func doOnce(ctx context.Context, method string, req OutboundRequest, timeout time.Duration) OutboundResult {
	attemptCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(attemptCtx, method, req.Url, bytes.NewReader(req.Body))
	if err != nil {
		return OutboundResult{Err: fmt.Errorf("new request: %w", err)}
	}
	httpReq.Header.Set("X-Custom-Header", "myvalue")
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range req.Headers {
		httpReq.Header.Add(k, fmt.Sprint(v))
	}
	trace := logging.TraceFromContext(ctx)
	if trace.RequestId != "" {
		if httpReq.Header.Get(HeaderRequestId) == "" {
			httpReq.Header.Set(HeaderRequestId, trace.RequestId)
		}
		if httpReq.Header.Get(HeaderParentId) == "" {
			httpReq.Header.Set(HeaderParentId, trace.RequestId)
		}
	}

	resp, err := outboundClient.Do(httpReq)
	if err != nil {
		if errors.Is(attemptCtx.Err(), context.DeadlineExceeded) && ctx.Err() == nil {
			err = fmt.Errorf("timeout after %s: %w", timeout, err)
		}
		return OutboundResult{Err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	result := OutboundResult{Body: body, StatusCode: resp.StatusCode, Status: resp.Status}
	if err != nil {
		result.Err = fmt.Errorf("read body: %w", err)
	}
	return result
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

// isRetryable ลองใหม่เฉพาะกรณีต่อไม่ติด (ยังไม่ได้ status) ไม่ retry 4xx / 5xx
func isRetryable(r OutboundResult) bool {
	return r.Err != nil && r.StatusCode == 0
}

// redactUrl ตัด query string และ user:password ออก เพราะอาจมีข้อมูลส่วนตัวหรือ key
func redactUrl(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		if i := strings.IndexByte(raw, '?'); i >= 0 {
			return raw[:i]
		}
		return raw
	}
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}
