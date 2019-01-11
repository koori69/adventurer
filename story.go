// Package adventurer @author K·J Create at 2019-01-09 11:04
package adventurer

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Story url路径映射
type Story struct {
	URL     string              `yaml:"url"`     // 请求的URL，支持正则表达式
	Method  []string            `yaml:"method"`  // 支持的请求方法
	Handler string              `yaml:"handler"` // 处理请求的方法名称，必须是对外公开的，即以大写开头
	Trials  map[string][]string `yaml:"trials"`  // 请求的预处理
}

// LoadStories load stories
func LoadStories(path string) (*[]Story, error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	var stories []Story
	err = yaml.Unmarshal(data, &stories)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	return &stories, nil
}
