# Adventurer

Go语言的路由器框架，支持URL正则匹配，API预处理等功能

## Features

- URL支持正则表达式
- API预处理
- URL支持指定请求Method
- 支持跨域

## Quick Start

### Story

使用yaml存储配置信息。

例：stories.yaml

```yaml
- url: "/test/hello" # 请求的URL
  method:  # 支持的Method方法，可以多个
    - "GET"
  handler: "DemoHandler"  # 处理struct 对象的handler方法名称，必须大写开头
  trials: # 预处理定义，map[string][]string类型，和Hook配对使用
    token:
      - "true"
    header:
      - "true"
    permission:
      - "ADMIN"
```

加载yaml信息

```go
stories, err := adventurer.LoadStories("./stories/stories.yaml")
if nil != err {
    log.Fatalln(err.Error())
}
```

### Hook

如果需要对URL进行预处理，需要实现Hook接口。

```go
type HeaderHook struct {
}

func NewHeaderHook() adventurer.Hook {
	return &HeaderHook{}
}

func (h *HeaderHook) Fire(prerequisite []string, equipment adventurer.Equipment) (bool, error) {
	if nil == prerequisite || len(prerequisite) < 1 {
		return false, nil
	}
	if "true" == prerequisite[0] {
		if "" == equipment.Header.Get("device") {
			return false, nil
		}
	}
	return true, nil
}
```

### 跨域

```go
adventurer.SetCros(true)
```



### Demo

```go
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	stories, err := adventurer.LoadStories("./stories/stories.yaml")
	if nil != err {
		log.Fatalln(err.Error())
	}
	hook := make(adventurer.StoryHook, 0)
	hook["header"] = NewHeaderHook()
	a, err := adventurer.NewAdventurer(Demo{}, stories,
		adventurer.NewProfile("/demo/about", "0.0.1", "1.11.2", "", "test"),
		&hook)
	if nil != err {
		log.Fatalln(err.Error())
	}
	http.HandleFunc("/", a.Explore)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 2111), nil))
}

// Demo demo
type Demo struct {
}

// DemoHandler demo handler
func (d *Demo) DemoHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}

type HeaderHook struct {
}

func NewHeaderHook() adventurer.Hook {
	return &HeaderHook{}
}

func (h *HeaderHook) Fire(prerequisite []string, equipment adventurer.Equipment) (bool, error) {
	if nil == prerequisite || len(prerequisite) < 1 {
		return false, nil
	}
	if "true" == prerequisite[0] {
		if "" == equipment.Header.Get("device") {
			return false, nil
		}
	}
	return true, nil
}
```

