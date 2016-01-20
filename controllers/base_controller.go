package controllers
import (
	"github.com/astaxie/beego"
	"beego_study/models"
	"beego_study/db"
	"beego_study/entities"
	"net/url"
	"strings"
	"encoding/json"
	"beego_study/exception"
	"io"
	"bytes"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
)

type BaseController struct {
	beego.Controller
}

type ResponseBody struct {
	Success bool
	Message interface{}
	Code    int
	Data    interface{}
}

func (c *BaseController) Prepare() {

	userId := c.CurrentUserId()
	beego.Error("userId:",userId)
	categories, _ := models.UserCategories(userId)
	c.Data["categories"] = categories
	c.Data["showLeftBar"] = true
	var keywords, _ = models.ParameterValue("index-keywords")
	c.Data["keywords"] = keywords
	var description, _ = models.ParameterValue("index-description")
	c.Data["description"] = description

	response := ResponseBody{Success:true}
	c.Data["response"] = response

	var args interface{}
	method := c.Ctx.Request.Method
	if ("GET" == method) {
		args = c.Ctx.Request.RequestURI
	}else {
		args = c.Ctx.Request.Form.Encode();
	}
	userAgent := c.Ctx.Request.UserAgent()
	userAgent = strings.ToLower(userAgent)
	//是否是手机访问
	c.Data["isMobile"] = false
	if (strings.Contains(userAgent, "android") || strings.Contains(userAgent, "iphone")) {
		c.Data["isMobile"] = true
	}
	user := c.CurrentUser()
	if nil != user {
		c.Data["user"] = user
	}

	beego.Info("request-params:", args)
}

func (c *BaseController) Finish() {

}

func (c *BaseController) Render() error {
	if !c.EnableRender {
		return nil
	}
	rb, err := c.RenderBytes()

	if err != nil {
		return err
	} else {
		c.Ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
		if err != nil {
			return err
		}
		miniRb, err := mini(rb)
		if err != nil {
			return err
		}
		c.Ctx.Output.Body(miniRb)
	}

	return nil
}

func mini(renderBytes []byte) ([]byte, error) {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		_, err := io.Copy(w, r)
		return err
	})
	m.AddFunc("text/javascript", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		_, err := io.Copy(w, r)
		return err
	})
	r := bytes.NewBuffer(renderBytes)
	w := &bytes.Buffer{}
	err := html.Minify(m, w, r, nil)

	return w.Bytes(), err
}


func (c *BaseController) NewPagination() *db.Pagination {
	page, err := c.GetInt("page")
	if nil != err {
		page = 1
	}
	pagination := db.NewPagination(page, 0, false)
	link, _ := url.ParseRequestURI(c.Ctx.Request.URL.String())
	pagination.SetUrl(link)
	return pagination
}


func (c *BaseController) CurrentUser() *entities.User {
	user := c.GetSession("user")
	if nil == user {
		return nil
	}
	var u, ok = user.(entities.User)
	if !ok {
		return nil
	}
	return &u
}

func (c *BaseController) CurrentOpenUser() *entities.OpenUser {
	openUser := c.GetSession("open_user")
	if nil == openUser {
		return nil
	}
	var u, ok = openUser.(entities.OpenUser)
	if !ok {
		return nil
	}
	return &u
}

func (c *BaseController) SetCurrSession(sessionKey string,value interface{})  {
	c.SetSession(sessionKey, value)
}


func (c *BaseController) CurrentUserId() int64 {
	user := c.CurrentUser()
	if nil == user {
		return 0
	}
	return user.Id
}
func (c *BaseController) StringError(message string) {
	response := new(ResponseBody)
	response.Code = -1
	response.Success = false
	response.Message = message
	result, err := json.Marshal(response)
	if nil == err {
		c.Data["message"] = string(result)
	}
}

func (c *BaseController) StringSuccess(message string) {
	response := new(ResponseBody)
	response.Code = 0
	response.Success = true
	response.Message = message
	result, err := json.Marshal(response)
	if nil == err {
		c.Data["message"] = string(result)
	}
}

func (c *BaseController) ErrorCodeJsonError(exception exception.ErrorCode) {
	response := new(ResponseBody)
	response.Code = exception.Code()
	response.Success = false
	response.Message = exception.Error()
	c.Data["json"] = response
	c.ServeJSON()
}


func (c *BaseController) JsonError(message interface{}) {
	response := new(ResponseBody)
	response.Code = -1
	response.Success = false
	response.Message = message
	c.Data["json"] = response
	c.ServeJSON()
}

func (c *BaseController) JsonSuccess(message interface{}) {
	response := new(ResponseBody)
	response.Code = 0
	response.Success = true
	response.Message = message
	c.Data["json"] = response
	c.ServeJSON()
}


func (c *BaseController) Ip() string {
	return c.Ctx.Request.Header.Get("X-Real-Ip")
}

func (c *BaseController) SetKeywords(keywords string) *BaseController {
	c.Data["keywords"] = keywords
	return c
}

func (c *BaseController) SetDescription(description string) *BaseController {
	c.Data["description"] = description
	return c
}

func (c *BaseController) SetTitle(title string) *BaseController {
	c.Data["title"] = title
	return c
}