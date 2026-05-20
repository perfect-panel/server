package hertzx

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"
)

type H map[string]interface{}

type HandlerFunc func(*Context)

type Context struct {
	Request *http.Request
	Writer  ResponseWriter
	Errors  ErrorChain

	base context.Context
	ctx  *app.RequestContext
	keys map[string]interface{}
}

const contextKey = "__ppanel_hertzx_context"

func NewContext(base context.Context, ctx *app.RequestContext) *Context {
	if value, ok := ctx.Get(contextKey); ok {
		if c, ok := value.(*Context); ok {
			return c
		}
	}
	c := &Context{
		base:    base,
		ctx:     ctx,
		Request: compatRequest(base, ctx),
		Writer:  newResponseWriter(ctx),
		keys:    make(map[string]interface{}),
	}
	ctx.Set(contextKey, c)
	return c
}

func Wrap(handler HandlerFunc) app.HandlerFunc {
	return func(base context.Context, ctx *app.RequestContext) {
		c := NewContext(base, ctx)
		handler(c)
		c.flush()
	}
}

func (c *Context) Next() {
	c.ctx.Next(c.Request.Context())
}

func (c *Context) Abort() {
	c.ctx.Abort()
}

func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Abort()
}

func (c *Context) ShouldBind(obj interface{}) error {
	if isJSONRequest(c.Request) {
		return c.ShouldBindJSON(obj)
	}
	if err := bindValues(obj, c.Request.URL.Query()); err != nil {
		return err
	}
	if len(c.ctx.Request.Body()) == 0 {
		return nil
	}
	return c.ctx.Bind(obj)
}

func (c *Context) ShouldBindJSON(obj interface{}) error {
	return c.ctx.BindJSON(obj)
}

func (c *Context) BindJSON(obj interface{}) error {
	return c.ShouldBindJSON(obj)
}

func (c *Context) ShouldBindQuery(obj interface{}) error {
	return bindValues(obj, c.Request.URL.Query())
}

func (c *Context) ShouldBindUri(obj interface{}) error {
	values := make(url.Values)
	for _, param := range c.ctx.Params {
		values.Add(param.Key, param.Value)
	}
	return bindValues(obj, values)
}

func (c *Context) Param(key string) string {
	return c.ctx.Param(key)
}

func (c *Context) Query(key string) string {
	return c.ctx.Query(key)
}

func (c *Context) GetQuery(key string) (string, bool) {
	return c.ctx.GetQuery(key)
}

func (c *Context) GetHeader(key string) string {
	return string(c.ctx.GetHeader(key))
}

func (c *Context) Header(key, value string) {
	c.Writer.Header().Set(key, value)
	c.ctx.Header(key, value)
}

func (c *Context) JSON(code int, obj interface{}) {
	if _, ok := c.Writer.(*responseWriter); !ok {
		data, err := json.Marshal(obj)
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Writer.WriteHeader(code)
		_, _ = c.Writer.Write(data)
		return
	}
	c.ctx.JSON(code, obj)
	if rw, ok := c.Writer.(*responseWriter); ok {
		rw.status = code
		rw.written = true
	}
}

func (c *Context) String(code int, format string, values ...interface{}) {
	if _, ok := c.Writer.(*responseWriter); !ok {
		body := fmt.Sprintf(format, values...)
		c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		c.Writer.WriteHeader(code)
		_, _ = c.Writer.WriteString(body)
		return
	}
	c.ctx.String(code, format, values...)
	if rw, ok := c.Writer.(*responseWriter); ok {
		rw.status = code
		rw.written = true
		rw.size += len(fmt.Sprintf(format, values...))
	}
}

func (c *Context) HTML(code int, name string, obj interface{}) {
	c.ctx.HTML(code, name, obj)
	if rw, ok := c.Writer.(*responseWriter); ok {
		rw.status = code
		rw.written = true
	}
}

func (c *Context) Redirect(code int, location string) {
	if _, ok := c.Writer.(*responseWriter); !ok {
		c.Writer.Header().Set("Location", location)
		c.Writer.WriteHeader(code)
		return
	}
	c.ctx.Redirect(code, []byte(location))
	if rw, ok := c.Writer.(*responseWriter); ok {
		rw.status = code
		rw.written = true
	}
}

func (c *Context) Status(code int) {
	if _, ok := c.Writer.(*responseWriter); !ok {
		c.Writer.WriteHeader(code)
		return
	}
	c.ctx.Status(code)
	if rw, ok := c.Writer.(*responseWriter); ok {
		rw.status = code
	}
}

func (c *Context) ClientIP() string {
	return c.ctx.ClientIP()
}

func (c *Context) FullPath() string {
	return c.ctx.FullPath()
}

func (c *Context) Set(key string, value interface{}) {
	c.keys[key] = value
	c.ctx.Set(key, value)
}

func (c *Context) Get(key string) (interface{}, bool) {
	if value, ok := c.keys[key]; ok {
		return value, true
	}
	return c.ctx.Get(key)
}

func (c *Context) GetString(key string) string {
	value, ok := c.Get(key)
	if !ok || value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return fmt.Sprint(value)
}

func (c *Context) Error(err error) *Error {
	if err == nil {
		return nil
	}
	item := &Error{Err: err}
	c.Errors = append(c.Errors, item)
	return item
}

func (c *Context) Deadline() (time.Time, bool) {
	return c.Request.Context().Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.Request.Context().Done()
}

func (c *Context) Err() error {
	return c.Request.Context().Err()
}

func (c *Context) Value(key interface{}) interface{} {
	if value := c.Request.Context().Value(key); value != nil {
		return value
	}
	if str, ok := key.(string); ok {
		if value, exists := c.Get(str); exists {
			return value
		}
	}
	return nil
}

func (c *Context) flush() {
	if c.Writer != nil {
		copyHeaders(c.ctx, c.Writer.Header())
	}
}

type Error struct {
	Err error
}

func (e *Error) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

type ErrorChain []*Error

func (e ErrorChain) Last() *Error {
	if len(e) == 0 {
		return nil
	}
	return e[len(e)-1]
}

func (e ErrorChain) String() string {
	var parts []string
	for _, err := range e {
		if err != nil && err.Err != nil {
			parts = append(parts, err.Err.Error())
		}
	}
	return strings.Join(parts, "\n")
}

type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier

	Status() int
	Size() int
	Written() bool
	WriteHeaderNow()
	WriteString(string) (int, error)
	Pusher() http.Pusher
}

type responseWriter struct {
	ctx     *app.RequestContext
	header  http.Header
	status  int
	size    int
	written bool
	mu      sync.Mutex
}

func newResponseWriter(ctx *app.RequestContext) *responseWriter {
	return &responseWriter{
		ctx:    ctx,
		header: make(http.Header),
		status: http.StatusOK,
	}
}

func (w *responseWriter) Header() http.Header {
	return w.header
}

func (w *responseWriter) WriteHeader(code int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.written {
		return
	}
	w.status = code
	w.written = true
	copyHeaders(w.ctx, w.header)
	w.ctx.SetStatusCode(code)
}

func (w *responseWriter) WriteHeaderNow() {
	w.WriteHeader(w.status)
}

func (w *responseWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ctx.Write(data)
	w.size += n
	return n, err
}

func (w *responseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) Written() bool {
	return w.written
}

func (w *responseWriter) Flush() {
	if flusher := w.ctx.GetWriter(); flusher != nil {
		_ = flusher.Flush()
	}
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn := w.ctx.GetConn()
	if netConn, ok := conn.(net.Conn); ok {
		return netConn, bufio.NewReadWriter(bufio.NewReader(netConn), bufio.NewWriter(netConn)), nil
	}
	return nil, nil, http.ErrNotSupported
}

func (w *responseWriter) CloseNotify() <-chan bool {
	ch := make(chan bool, 1)
	go func() {
		<-w.ctx.Finished()
		ch <- true
	}()
	return ch
}

func (w *responseWriter) Pusher() http.Pusher {
	return nil
}

func copyHeaders(ctx *app.RequestContext, headers http.Header) {
	for key, values := range headers {
		for _, value := range values {
			ctx.Response.Header.Add(key, value)
		}
	}
}

func compatRequest(base context.Context, ctx *app.RequestContext) *http.Request {
	body := ctx.Request.Body()
	req, err := http.NewRequestWithContext(base, string(ctx.Method()), ctx.URI().String(), bytes.NewReader(body))
	if err != nil {
		req, _ = http.NewRequestWithContext(base, string(ctx.Method()), "/", bytes.NewReader(body))
	}
	req.Header = make(http.Header)
	ctx.Request.Header.VisitAll(func(k, v []byte) {
		req.Header.Add(string(k), string(v))
	})
	req.Host = string(ctx.Host())
	req.RemoteAddr = ctx.RemoteAddr().String()
	req.RequestURI = string(ctx.URI().RequestURI())
	req.ContentLength = int64(len(body))
	return req
}

func isJSONRequest(req *http.Request) bool {
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return strings.Contains(contentType, "json")
	}
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

func bindValues(dst interface{}, values url.Values) error {
	if dst == nil {
		return nil
	}
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("bind target must be a non-nil pointer")
	}
	return bindValue(v.Elem(), values)
}

func bindValue(v reflect.Value, values url.Values) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return bindValue(v.Elem(), values)
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		sf := t.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous {
			continue
		}
		if sf.Anonymous {
			if err := bindValue(field, values); err != nil {
				return err
			}
			continue
		}
		name := fieldName(sf)
		if name == "" {
			continue
		}
		raw, ok := values[name]
		if !ok && field.Kind() == reflect.Slice {
			raw, ok = values[name+"[]"]
		}
		if !ok || len(raw) == 0 {
			continue
		}
		if err := setField(field, raw); err != nil {
			return fmt.Errorf("bind %s: %w", sf.Name, err)
		}
	}
	return nil
}

func fieldName(sf reflect.StructField) string {
	for _, key := range []string{"form", "query", "uri", "path", "json"} {
		tag := sf.Tag.Get(key)
		if tag == "-" {
			return ""
		}
		if tag != "" {
			return strings.Split(tag, ",")[0]
		}
	}
	return sf.Name
}

func setField(field reflect.Value, raw []string) error {
	if !field.CanSet() {
		return nil
	}
	if field.Kind() == reflect.Ptr {
		if len(raw) == 0 || raw[0] == "" {
			return nil
		}
		field.Set(reflect.New(field.Type().Elem()))
		return setField(field.Elem(), raw)
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(raw[0])
	case reflect.Bool:
		value, err := strconv.ParseBool(raw[0])
		if err != nil {
			return err
		}
		field.SetBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(raw[0], 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err := strconv.ParseUint(raw[0], 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(value)
	case reflect.Float32, reflect.Float64:
		value, err := strconv.ParseFloat(raw[0], field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(value)
	case reflect.Slice:
		slice := reflect.MakeSlice(field.Type(), 0, len(raw))
		for _, item := range raw {
			elem := reflect.New(field.Type().Elem()).Elem()
			if err := setField(elem, []string{item}); err != nil {
				return err
			}
			slice = reflect.Append(slice, elem)
		}
		field.Set(slice)
	case reflect.Struct:
		return nil
	}
	return nil
}

type Engine struct {
	h               *server.Hertz
	RemoteIPHeaders []string
}

func New(opts ...config.Option) *Engine {
	return &Engine{h: server.New(opts...)}
}

func Default(opts ...config.Option) *Engine {
	return &Engine{h: server.Default(opts...)}
}

func (e *Engine) Hertz() *server.Hertz {
	return e.h
}

func (e *Engine) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{group: e.h.Group(relativePath, wrapHandlers(handlers)...)}
}

func (e *Engine) Use(handlers ...HandlerFunc) {
	e.h.Use(wrapHandlers(handlers)...)
}

func (e *Engine) GET(path string, handlers ...HandlerFunc) {
	e.h.GET(path, wrapHandlers(handlers)...)
}

func (e *Engine) POST(path string, handlers ...HandlerFunc) {
	e.h.POST(path, wrapHandlers(handlers)...)
}

func (e *Engine) PUT(path string, handlers ...HandlerFunc) {
	e.h.PUT(path, wrapHandlers(handlers)...)
}

func (e *Engine) DELETE(path string, handlers ...HandlerFunc) {
	e.h.DELETE(path, wrapHandlers(handlers)...)
}

func (e *Engine) Any(path string, handlers ...HandlerFunc) {
	wrapped := wrapHandlers(handlers)
	e.h.GET(path, wrapped...)
	e.h.POST(path, wrapped...)
	e.h.PUT(path, wrapped...)
	e.h.DELETE(path, wrapped...)
	e.h.PATCH(path, wrapped...)
	e.h.OPTIONS(path, wrapped...)
	e.h.HEAD(path, wrapped...)
}

func (e *Engine) NoRoute(handlers ...HandlerFunc) {
	e.h.NoRoute(wrapHandlers(handlers)...)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := app.NewContext(0)
	_ = adaptor.CopyToHertzRequest(r, &ctx.Request)
	e.h.ServeHTTP(r.Context(), ctx)
	ctx.Response.Header.VisitAll(func(k, v []byte) {
		w.Header().Add(string(k), string(v))
	})
	status := ctx.Response.StatusCode()
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = w.Write(ctx.Response.Body())
}

func (e *Engine) Run(addr ...string) error {
	if len(addr) > 0 && addr[0] != "" {
		e.h.GetOptions().Addr = addr[0]
	}
	return e.h.Run()
}

func (e *Engine) RunTLS(addr, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	if addr != "" {
		e.h.GetOptions().Addr = addr
	}
	e.h.GetOptions().TLS = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	}
	return e.h.Run()
}

func (e *Engine) LoadHTMLGlob(pattern string) {
	e.h.LoadHTMLGlob(pattern)
}

func (e *Engine) LoadHTMLFiles(files ...string) {
	e.h.LoadHTMLFiles(files...)
}

func (e *Engine) SetHTMLTemplate(tmpl *template.Template) {
	e.h.SetHTMLTemplate(tmpl)
}

type RouterGroup struct {
	group *route.RouterGroup
}

func (g *RouterGroup) Use(handlers ...HandlerFunc) {
	g.group.Use(wrapHandlers(handlers)...)
}

func (g *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{group: g.group.Group(relativePath, wrapHandlers(handlers)...)}
}

func (g *RouterGroup) GET(path string, handlers ...HandlerFunc) {
	g.group.GET(path, wrapHandlers(handlers)...)
}

func (g *RouterGroup) POST(path string, handlers ...HandlerFunc) {
	g.group.POST(path, wrapHandlers(handlers)...)
}

func (g *RouterGroup) PUT(path string, handlers ...HandlerFunc) {
	g.group.PUT(path, wrapHandlers(handlers)...)
}

func (g *RouterGroup) DELETE(path string, handlers ...HandlerFunc) {
	g.group.DELETE(path, wrapHandlers(handlers)...)
}

func (g *RouterGroup) Any(path string, handlers ...HandlerFunc) {
	wrapped := wrapHandlers(handlers)
	g.group.GET(path, wrapped...)
	g.group.POST(path, wrapped...)
	g.group.PUT(path, wrapped...)
	g.group.DELETE(path, wrapped...)
	g.group.PATCH(path, wrapped...)
	g.group.OPTIONS(path, wrapped...)
	g.group.HEAD(path, wrapped...)
}

func wrapHandlers(handlers []HandlerFunc) []app.HandlerFunc {
	if len(handlers) == 0 {
		return nil
	}
	wrapped := make([]app.HandlerFunc, 0, len(handlers))
	for _, handler := range handlers {
		wrapped = append(wrapped, Wrap(handler))
	}
	return wrapped
}

func MethodPost(c *Context) bool {
	return string(c.ctx.Method()) == consts.MethodPost
}

func SyncRequestBody(c *Context) {
	if c.Request == nil || c.Request.Body == nil {
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return
	}
	c.ctx.Request.SetBody(body)
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	c.Request.ContentLength = int64(len(body))
}

func SyncRequestURI(c *Context) {
	if c.Request == nil || c.Request.URL == nil {
		return
	}
	uri := c.Request.URL.RequestURI()
	c.ctx.Request.SetRequestURI(uri)
}

func RequestContext(c *Context) *app.RequestContext {
	return c.ctx
}

func HertzRequest(c *Context) *protocol.Request {
	return &c.ctx.Request
}

const ReleaseMode = "release"

func SetMode(string) {}

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if value := recover(); value != nil {
				c.String(http.StatusInternalServerError, "Internal Server Error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
