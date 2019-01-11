// Package adventurer @author K·J Create at 2019-01-09 11:00
package adventurer

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var logger = DefaultLog()

// Adventurer 路由对象
type Adventurer struct {
	owner   interface{}
	stories []Story
	profile *Profile
	cros    bool
	hook    *StoryHook
}

// Equipment 请求的信息
type Equipment struct {
	Query  url.Values
	Body   []byte
	Header http.Header
	Method string
}

// Hook 校验
type Hook interface {
	Fire(prerequisite []string, equipment Equipment) (bool, error)
}

// StoryHook
//type StoryHook struct {
//	Info map[string]int // 基本信息，key：校验的名字，value：校验的优先级，数字越小优先级越高
//	Hook Hook           // 校验
//}

type StoryHook map[string]Hook // 基本信息，key：校验的名字，value：校验实现

// NewAdventurer 生成路由对象
func NewAdventurer(owner interface{}, stories *[]Story, profile *Profile, hook *StoryHook) (*Adventurer, error) {
	if nil == owner || nil == stories {
		logger.Error("param is nil")
		return nil, errors.New("param is nil")
	}
	adventurer := &Adventurer{owner: owner, profile: profile, hook: hook}
	err := adventurer.InitStory(*stories)
	if nil != err {
		logger.Error(err.Error())
		return nil, err
	}
	if nil != profile || "" != profile.URL {
		story := Story{
			URL:     profile.URL,
			Handler: "Handler",
			Method:  []string{http.MethodGet},
		}
		err = adventurer.AddStory(story)
		if nil != err {
			logger.Error(err.Error())
			return nil, err
		}
	}
	return adventurer, nil
}

// AddStory add story
func (a *Adventurer) AddStory(s Story) error {
	if "" == s.URL || nil == s.Method || "" == s.Handler {
		logger.Error("param is invalid")
		return errors.New("param is invalid")
	}
	methods := strings.Join(s.Method, ",")
	if nil == a.stories {
		a.stories = make([]Story, 0)
	}
	for _, v := range a.stories {
		if ok, _ := regexp.MatchString("[A-Z].*", v.Handler); !ok {
			logger.Error("handler should be exported")
			return errors.New("handler should be exported")
		}
		if v.URL == s.URL {
			for _, m := range v.Method {
				if strings.Contains(methods, m) {
					logger.Error("url handler already exist")
					return errors.New("url handler already exist")
				}
			}
		}
	}
	a.stories = append(a.stories, s)
	return nil
}

// InitStory 初始化story
func (a *Adventurer) InitStory(stories []Story) error {
	if nil == stories {
		logger.Error("stories is nil")
		return errors.New("stories is nil")
	}
	for _, s := range stories {
		err := a.AddStory(s)
		if nil != err {
			logger.Error(err.Error())
			return err
		}
	}
	return nil
}

// SetCros enable or disable cros
func (a *Adventurer) SetCros(cros bool) {
	a.cros = cros
}

// Explore url router
func (a *Adventurer) Explore(w http.ResponseWriter, r *http.Request) {
	startTime := unixMillisecond()
	if "" == r.Header.Get("X-Real-IP") {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		r.Header.Set("X-Real-IP", ip)
	}
	if a.cros {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
	}
	found := false
	match := false
	for _, v := range a.stories {
		methods := strings.Join(v.Method, ",")
		if ok, _ := regexp.MatchString("^"+v.URL+"$", r.URL.Path); ok {
			found = true
			if strings.Contains(methods, r.Method) {
				match = true
				if nil != a.profile && r.URL.Path == a.profile.URL {
					a.profile.Handler(w, r)
					goto End
				}

				// 校验
				if nil != a.hook && len(*a.hook) > 0 {
					if nil != v.Trials {
						equipment, err := parseEquipment(r)
						if nil != err {
							logger.Error(err.Error())
							goto End
						}
						for k, t := range v.Trials {
							if trial, ok := (*a.hook)[k]; ok {
								if nil != trial {
									b, err := trial.Fire(t, *equipment)
									if nil != err {
										logger.Error(err.Error())
										goto End
									}
									if !b {
										w.WriteHeader(http.StatusBadRequest)
										goto End
									}
								}
							}
						}
					}
				}

				param := make([]reflect.Value, 2)
				param[0] = reflect.ValueOf(w)
				param[1] = reflect.ValueOf(r)
				reflect.ValueOf(a.owner).MethodByName(v.Handler).Call(param)
				break
			}
		}
	}
End:
	endTime := unixMillisecond()
	logger.WithFields(logrus.Fields{
		"url":  r.URL.Path,
		"cost": fmt.Sprintf("%d ms", endTime-startTime),
	}).Info("OK")
	if found && !match {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

// parseEquipment 获取request
func parseEquipment(r *http.Request) (*Equipment, error) {
	switch r.Method {
	case http.MethodGet:
		return &Equipment{Header: r.Header, Query: r.URL.Query(), Method: r.Method}, nil
	case http.MethodDelete:
		fallthrough
	case http.MethodPost:
		fallthrough
	case http.MethodPut:
		exp, err := regexp.Compile("([a-zA-z/\\-_\\.]+)(;*(.+))*")
		if nil != err {
			logger.Error(err.Error())
			return nil, err
		}
		param := exp.FindStringSubmatch(r.Header.Get("Content-Type"))
		if nil == param || len(param) < 2 {
			logger.Error("not found content-type")
			return nil, errors.New("not found content-type")
		}
		switch param[1] {
		case "multipart/form-data":
			body, err := httputil.DumpRequest(r, true)
			if nil != err {
				logger.Error(err.Error())
				return nil, err
			}
			copy, err := http.NewRequest("POST", "", bytes.NewReader(body))
			if nil != err {
				logger.Error(err.Error())
				return nil, err
			}
			copy.Header = r.Header
			err = copy.ParseMultipartForm(32 << 20)
			if nil != err {
				copy.Body.Close()
				logger.Error(err.Error())
				return nil, err
			}
			copy.Body.Close()
			return &Equipment{Header: r.Header, Query: r.PostForm, Method: r.Method}, nil
		//case "application/json":
		//	body, err := ioutil.ReadAll(r.Body)
		//	if nil != err {
		//		logger.Error(err.Error())
		//		return nil, err
		//	}
		//	r.Body.Close()
		//	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		//	return &Equipment{Header: r.Header, Query: r.URL.Query(), Method: r.Method, Body: body}, nil
		default:
			body, err := ioutil.ReadAll(r.Body)
			if nil != err {
				logger.Error(err.Error())
				return nil, err
			}
			r.Body.Close()
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			return &Equipment{Header: r.Header, Query: r.URL.Query(), Method: r.Method, Body: body}, nil
		}
	}
	return nil, errors.New("method not supported")
}

// unixMillisecond unix time millisecond
func unixMillisecond() int64 {
	return time.Now().UnixNano() / 1e6
}
