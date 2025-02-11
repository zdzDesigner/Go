package tt

import (
	"net/http/httptest"
	"time"
	"battery/src/db/mysql"
	"battery/src/util/ginc"

	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gopkg.in/resty.v1"
)

// Tester ..
type Tester interface {
	SetHeader(map[string]string) Tester
	SetQuery(map[string]string) Tester
	Post(any) (*resty.Response, error)
	Del(any) (*resty.Response, error)
	Put(params any) (*resty.Response, error)
	Get() (*resty.Response, error)
}

// Test ..
type Test struct {
	Cli *resty.Client
	api string
}

// SetHeader ..
func (t *Test) SetHeader(header map[string]string) Tester {
	for k, v := range header {
		t.Cli = t.Cli.SetHeader(k, v)
	}
	return t
}

// SetQuery ..
func (t *Test) SetQuery(query map[string]string) Tester {
	t.Cli = t.Cli.SetQueryParams(query)
	return t
}
func (t *Test) Get() (*resty.Response, error) {
	return t.Cli.R().Get("")
}

// Post ..
func (t *Test) Post(params any) (*resty.Response, error) {
	return t.Cli.R().SetBody(params).Post("")
}
func (t *Test) Del(params any) (*resty.Response, error) {
	return t.Cli.R().SetBody(params).Delete((""))
}

// Put ..
func (t *Test) Put(params any) (*resty.Response, error) {
	return t.Cli.R().SetBody(params).Put("")
}

// NewTest ..
func NewTest(hook func(app *gin.Engine) *gin.Engine) Tester {
	var (
		cli       *resty.Client
		timelocal = time.FixedZone("CST", 3600*8)
	)
	time.Local = timelocal

	gin.SetMode(gin.TestMode)
	if cli == nil {
		app := hook(gin.New())
		server := httptest.NewServer(app)
		Plugins()
		cli = resty.New().SetHostURL(server.URL)
	}

	return &Test{
		Cli: cli,
	}
}

type HookFunc func(*gin.Engine) *gin.Engine

func (hf HookFunc) Use(hook HookFunc) {
	// hook.Use()

}

// type HookFunc func(*gin.Context) *gin.Context
type Hook struct {
	app *gin.Engine
}

func (h *Hook) Use(hook HookFunc) *Hook {
	hook(h.app)
	return h
}
func (h *Hook) Route(routeHandle func(ginc.Contexter)) *Hook {
	h.app.Use(ginc.Hander(routeHandle))
	return h
}

func THandler(routeHandle func(ginc.Contexter), hooks ...HookFunc) Tester {
	return NewTest(func(app *gin.Engine) *gin.Engine {
		lo.ForEach(hooks, func(hook HookFunc, _ int) { hook(app) })
		app.Use(func(ctx *gin.Context) {
			ctx.Next()
		})

		app.Use(ginc.Hander(routeHandle))
		return app
	})
}

// TODO:: hook prototype
func THook(app *gin.Engine) func(func(ginc.Contexter)) {
	return func(func(ginc.Contexter)) {
	}
}

func TJsonHandler(routeHandle func(ginc.Contexter), hooks ...HookFunc) Tester {
	return THandler(routeHandle, hooks...).SetHeader(map[string]string{
		"Content-Type": "application/json",
		// "x-forwarded-for": "219.133.168.5",
	})
}

func JsonBody(resp *resty.Response, err error) (*resty.Response, int, error) {
	bresp, _ := simplejson.NewJson(resp.Body())
	code, _ := bresp.Get("code").Int()

	return resp, code, err
}

func Plugins() {
	mysql.Boot()
	// go brand.Start(url)
	// go storage.Start()
	// go dynamic.Start()
}
