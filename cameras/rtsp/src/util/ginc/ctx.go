package ginc

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	// "strconv"
	"strings"
	"sync"
	"battery/src/util/lang"

	"github.com/gin-gonic/gin"
	"gopkg.in/resty.v1"
)

// Contexter wh 接口
type Contexter interface {
	ParseReqbody(any) error                 // 解析body入参
	Success(...map[string]any)              // 标识成功返回状态
	SuccessByte(b []byte, params ...string) // 原生数据
	Fail(any)                               // 标识失败返回状态
	FailErr(int, string)                    // 标识失败返回状态
	GinCtx() *gin.Context                   // 获取gin context
	Query(string) string                    // 获取query
	ClientPost(api string, params any, options ...map[string]string) (
		*resty.Response, error) // Post请求
	ClientPut(api string, params any, options ...map[string]string) (
		*resty.Response, error) // Post请求
	ClientGet(api string, params any, options ...map[string]string) (
		*resty.Response, error) // Get 请求
	ParamRoute(string) string // 路由参数解析
	Log(string)               // 记录log
	Logging(string)           // 不换行合并

}

// Context ...
type Context struct {
	Gin      *gin.Context
	Client   *Client
	keys     map[string]any
	m        sync.Mutex
	logs     []string
	loggings []string
	tally    map[string]string
}

// Hander ..
func Hander(fn func(c Contexter)) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := &Context{Gin: c}
		fn(ctx)
	}
}

// GinCtx gin context
func (c *Context) GinCtx() *gin.Context {
	return c.Gin
}
func (c *Context) Query(query string) string {
	return c.Gin.Query(query)
}

// ParamRoute ..
func (c *Context) ParamRoute(key string) string {
	return strings.Replace(c.Gin.Param("id"), "/", "", -1)
}

// ParseReqbody 解析网络body
func (c *Context) ParseReqbody(reqbody any) error {
	c.ReWrite()
	// fmt.Println("c.Gin Body:", c.Gin.Request.Body)
	if err := c.Gin.ShouldBindJSON(reqbody); err != nil && err != io.EOF {
		c.Fail(map[string]any{
			"code":   40000,
			"errmsg": fmt.Sprintf("参数解析 :%s", err.Error()),
		})
		return err
	}
	return nil
}

// Success 成功
func (c *Context) Success(datas ...map[string]any) {
	data := make(map[string]any)
	if len(datas) > 0 {
		data = datas[0]
	}
	if data["code"] == nil {
		data["code"] = 0
	}

	// code, _ := data["code"].(int)
	c.Log(fmt.Sprintf("code:%d", data["code"]))
	c.ReWrite()
	if b, err := json.Marshal(data); err != nil {
		panic(err)
	} else {
		c.Gin.Data(200, "application/json", b)
	}
}

// SuccessByte 成功
func (c *Context) SuccessByte(b []byte, params ...string) {
	c.ReWrite()
	contentType := append(params, "application/json")[0]
	c.Gin.Data(200, contentType, b)

}

// Fail 失败
func (c *Context) Fail(data any) {
	c.ReWrite()
	if b, err := json.Marshal(data); err != nil {
		panic(err)
	} else {
		c.Gin.Data(200, "application/json", b)
	}
}

func (c *Context) FailErr(code int, errmsg string) {
	c.ReWrite()
	c.Fail(map[string]any{
		"code":   code,
		"errmsg": errmsg,
	})
}

func (c *Context) ReWrite() {
	c.Gin.Set("LOGIC_LIST", c.logs)
}

func (c *Context) Logging(log string) {
	c.m.Lock()
	defer c.m.Unlock()
	if c.loggings == nil {
		c.loggings = make([]string, 0)
	}
	c.loggings = append(c.loggings, log)
}
func (c *Context) Log(log string) {
	c.m.Lock()
	defer c.m.Unlock()
	if c.logs == nil {
		c.logs = make([]string, 0)
	}

	c.logs = append(c.logs, strings.Join(append(c.loggings, log), ""))
	c.loggings = []string{}
}

// rget rewrite Gin
func (c *Context) rget(key string) (value any, exists bool) {
	value, exists = c.keys[key]
	return
}

// rset rewrite Gin
func (c *Context) rset(key string, value any) {
	if c.keys == nil {
		c.keys = make(map[string]any)
	}
	c.keys[key] = value
}

// rgetint rewrite Gin
func (c *Context) rgetint(key string) (i int) {
	if val, ok := c.rget(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// ClientPost method
func (c *Context) ClientPost(api string, params any, options ...map[string]string) (
	*resty.Response, error) {
	return c.request(c.Client.PostRow)(api, params, options...)
}

func (c *Context) ClientPut(api string, params any, options ...map[string]string) (
	*resty.Response, error) {
	return c.request(c.Client.PutRow)(api, params, options...)
}

// ClientGet method
func (c *Context) ClientGet(api string, params any, options ...map[string]string) (
	*resty.Response, error) {
	return c.request(c.Client.GetRow)(api, params, options...)
}

type requestHander func(string, any, ...map[string]string) (
	*resty.Response, map[string]any, error)

type requestFunc func(string, any, ...map[string]string) (
	*resty.Response, error)

// request 公共部分(锁操作)
func (c *Context) request(fn requestHander) requestFunc {
	return func(api string, params any, options ...map[string]string) (*resty.Response, error) {
		// fmt.Printf("options===>%+v\n", options)

		// 内容服务传递监控headers
		if strings.Contains(api, os.Getenv("KONG_CONTENT_SERVER_INTERNAL")) {
			headers := c.Gin.GetStringMapString("headers")
			if headers != nil {
				if len(options) == 1 {
					options = []map[string]string{lang.MapAssign(headers, options[0])}
				} else {
					options = []map[string]string{headers}
				}
			}
		}

		// 过滤正版资源
		params = c.fillSourceFilter(params)
		resp, log, err := fn(api, params, options...)

		c.m.Lock()
		total := c.rgetint("REQUEST_COUNT")
		if total == 0 {
			c.rset("REQUEST_COUNT", 1)
		} else {
			c.rset("REQUEST_COUNT", total+1)
		}
		if sourceTemp, ok := c.rget("REQUEST_SOURCE"); ok {
			source := sourceTemp.([]map[string]any)
			source = append(source, log)
			c.rset("REQUEST_SOURCE", source)
		} else {
			c.rset("REQUEST_SOURCE", []map[string]any{log})
		}

		c.m.Unlock()
		return resp, err
	}
}

func (c *Context) fillSourceFilter(params any) any {

	sourceFilter := c.Gin.GetString("sourceFilter")
	if sourceFilter == "1" {
		if newparams, ok := params.(map[string]string); ok {
			newparams["sourceFilter"] = sourceFilter
			return newparams
		}
		if newparams, ok := params.(map[string]any); ok {
			newparams["sourceFilter"] = sourceFilter
			return newparams
		}
		// if b, err := json.Marshal(params); err == nil {
		// 	fmt.Println(string(b))
		// }
	}
	return params
}

// NewContext ..
func NewContext() Contexter {
	return &Context{Gin: &gin.Context{}}
}
