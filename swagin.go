package swagin

import (
	"embed"
	"encoding/json"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/long2ice/swagin/router"
	"github.com/long2ice/swagin/swagger"
	"html/template"
	"net/http"
)

//go:embed templates/*
var templates embed.FS

type SwaGin struct {
	*gin.Engine
	Swagger  *swagger.Swagger
	Routers  map[string]map[string]*router.Router
	subApps  map[string]*SwaGin
	rootPath string
}

func New(swagger *swagger.Swagger) *SwaGin {
	f := &SwaGin{Engine: gin.New(), Swagger: swagger, Routers: make(map[string]map[string]*router.Router), subApps: make(map[string]*SwaGin)}
	f.SetHTMLTemplate(template.Must(template.ParseFS(templates, "templates/*.html")))
	if swagger != nil {
		swagger.Routers = f.Routers
	}
	return f
}

func NewFromEngine(engine *gin.Engine, swagger *swagger.Swagger) *SwaGin {
	f := &SwaGin{Engine: engine, Swagger: swagger, Routers: make(map[string]map[string]*router.Router), subApps: make(map[string]*SwaGin)}
	f.SetHTMLTemplate(template.Must(template.ParseFS(templates, "templates/*.html")))
	if swagger != nil {
		swagger.Routers = f.Routers
	}
	return f
}

func (g *SwaGin) Mount(path string, app *SwaGin) {
	app.rootPath = path
	app.Engine = g.Engine
	app.Swagger.Servers = append(app.Swagger.Servers, &openapi3.Server{
		URL: path,
	})
	g.subApps[path] = app
}

func (g *SwaGin) Group(path string, options ...Option) *Group {
	group := &Group{
		SwaGin: g,
		Path:   path,
	}
	for _, option := range options {
		option(group)
	}
	return group
}

func (g *SwaGin) Handle(path string, method string, r *router.Router) {
	r.Method = method
	r.Path = path
	if g.Routers[path] == nil {
		g.Routers[path] = make(map[string]*router.Router)
	}
	g.Routers[path][method] = r
}

func (g *SwaGin) GET(path string, router *router.Router) {
	g.Handle(path, http.MethodGet, router)
}

func (g *SwaGin) POST(path string, router *router.Router) {
	g.Handle(path, http.MethodPost, router)
}

func (g *SwaGin) HEAD(path string, router *router.Router) {
	g.Handle(path, http.MethodHead, router)
}

func (g *SwaGin) PATCH(path string, router *router.Router) {
	g.Handle(path, http.MethodPatch, router)
}

func (g *SwaGin) DELETE(path string, router *router.Router) {
	g.Handle(path, http.MethodDelete, router)
}

func (g *SwaGin) PUT(path string, router *router.Router) {
	g.Handle(path, http.MethodPut, router)
}

func (g *SwaGin) OPTIONS(path string, router *router.Router) {
	g.Handle(path, http.MethodOptions, router)
}

func (g *SwaGin) init() {
	g.initRouters()
	if g.Swagger == nil {
		return
	}
	g.Engine.GET(g.fullPath(g.Swagger.OpenAPIUrl), func(c *gin.Context) {
		c.JSON(http.StatusOK, g.Swagger)
	})
	g.Engine.GET(g.fullPath(g.Swagger.DocsUrl), func(c *gin.Context) {
		options := `{}`
		if g.Swagger.SwaggerOptions != nil {
			data, err := json.Marshal(g.Swagger.SwaggerOptions)
			if err != nil {
				panic(err)
			}
			options = string(data)
		}
		c.HTML(http.StatusOK, "swagger.html", gin.H{
			"openapi_url":     g.fullPath(g.Swagger.OpenAPIUrl),
			"title":           g.Swagger.Title,
			"swagger_options": options,
		})
	})
	g.Engine.GET(g.fullPath(g.Swagger.RedocUrl), func(c *gin.Context) {
		options := `{}`
		if g.Swagger.RedocOptions != nil {
			data, err := json.Marshal(g.Swagger.RedocOptions)
			if err != nil {
				panic(err)
			}
			options = string(data)
		}
		c.HTML(http.StatusOK, "redoc.html", gin.H{
			"openapi_url":   g.fullPath(g.Swagger.OpenAPIUrl),
			"title":         g.Swagger.Title,
			"redoc_options": options,
		})
	})
	g.Engine.GET(g.fullPath(g.Swagger.RapiDocUrl), func(c *gin.Context) {
		options := `{}`
		if g.Swagger.RapiDocOptions != nil {
			data, err := json.Marshal(g.Swagger.RapiDocOptions)
			if err != nil {
				panic(err)
			}
			options = string(data)
		}
		c.HTML(http.StatusOK, "rapidoc.html", gin.H{
			"openapi_url":     g.fullPath(g.Swagger.OpenAPIUrl),
			"title":           g.Swagger.Title,
			"rapidoc_options": options,
		})
	})
	g.Swagger.BuildOpenAPI()
}
func (g *SwaGin) initRouters() {
	for path, m := range g.Routers {
		path = g.fullPath(path)
		for method, r := range m {
			handlers := r.GetHandlers()
			if method == http.MethodGet {
				g.Engine.GET(path, handlers...)
			} else if method == http.MethodPost {
				g.Engine.POST(path, handlers...)
			} else if method == http.MethodHead {
				g.Engine.HEAD(path, handlers...)
			} else if method == http.MethodPatch {
				g.Engine.PATCH(path, handlers...)
			} else if method == http.MethodDelete {
				g.Engine.DELETE(path, handlers...)
			} else if method == http.MethodPut {
				g.Engine.PUT(path, handlers...)
			} else if method == http.MethodOptions {
				g.Engine.OPTIONS(path, handlers...)
			} else {
				g.Engine.Any(path, handlers...)
			}
		}
	}
}
func (g *SwaGin) fullPath(path string) string {
	return g.rootPath + path
}
func (g *SwaGin) Run(addr ...string) error {
	g.init()
	for _, s := range g.subApps {
		s.init()
	}
	return g.Engine.Run(addr...)
}
