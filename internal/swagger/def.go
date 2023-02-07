package swagger

var version = "2.0"
var reqCtxType = []string{"application/json"}
var respCtxType = []string{"application/json"}

// Head Swagger 版本
type Head struct {
	Swagger string `json:"swagger"`
}

// Info 指定 API 的 相关信息
type Info struct {
	Description string `json:"description"`
	Version     string `json:"version"`
	Title       string `json:"title"`
}

// ExternalDocs 外部文档
type ExternalDocs struct {
	Description string `json:"description,omitempty"` // 描述
	URL         string `json:"url,omitempty"`         // 外部文档地址
}

// Tag 标签
type Tag struct {
	Name         string        `json:"name"`                   // 名称
	Description  string        `json:"description"`            // 描述
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"` // 外部链接
}

// Schema 引用
type Schema struct {
	Ref string `json:"$ref,omitempty"` // 主体模式和响应主体模式中引用
}

// Element 元素定义
type Element struct {
	In          string  `json:"in"`                // 入参
	Name        string  `json:"name"`              // 参数名字
	Description string  `json:"description"`       // 描述
	Required    bool    `json:"required"`          // 是否必须
	Type        string  `json:"type,omitempty"`    // 类型
	Schema      *Schema `json:"schema,omitempty"`  // 引用
	Default     string  `json:"default,omitempty"` // 默认值
}

type Param struct {
	Tags        []string        `json:"tags"`                  // 分组标记
	Summary     string          `json:"summary"`               // 摘要
	Description string          `json:"description"`           // 描述
	OperationID string          `json:"operationId,omitempty"` // 操作id
	Consumes    []string        `json:"consumes"`              // 请求 content type
	Produces    []string        `json:"produces"`              // 响应 content type
	Parameters  []Element       `json:"parameters"`            // 请求参数
	Responses   map[string]Resp `json:"responses"`             // 返回参数
	Security    interface{}     `json:"security,omitempty"`    // 认证信息
	Deprecated  bool            `json:"deprecated,omitempty"`  // API是否过时
}

type Resp struct {
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
}

type SecurityDefinitions struct {
	Type string `json:"type"`
	Name string `json:"name"`
	In   string `json:"in"`
}

type PropertyItems struct {
	Type        string            `json:"type,omitempty"`
	Format      string            `json:"format,omitempty"`
	Description string            `json:"description,omitempty"` // 描述
	Enum        interface{}       `json:"enum,omitempty"`        // enum
	Items       map[string]string `json:"items,omitempty"`
	Ref         string            `json:"$ref,omitempty"` // 主体模式和响应主体模式中引用
}

type Property struct {
	Type        string         `json:"type,omitempty"` // 类型
	Items       *PropertyItems `json:"items,omitempty"`
	Format      string         `json:"format,omitempty"`      // format 类型
	Description string         `json:"description,omitempty"` // 描述
	Enum        []interface{}  `json:"enum,omitempty"`        // enum
	Ref         string         `json:"$ref,omitempty"`        // 主体模式和响应主体模式中引用
}

// XML xml
type XML struct {
	Name    string `json:"name"`
	Wrapped bool   `json:"wrapped"`
}

// Definition 通用结构体定义
type Definition struct {
	Type       string              `json:"type"`                 // 类型 object
	Properties map[string]Property `json:"properties,omitempty"` // 属性列表
	Items      map[string]Property `json:"items,omitempty"`

	//XML        XML                 `json:"xml"`
}

// APIBody swagger api body info
type APIBody struct {
	Head
	Info                Info                        `json:"info"`
	Host                string                      `json:"host"`     // http host
	BasePath            string                      `json:"basePath"` // 根级别
	Tags                []Tag                       `json:"tags"`
	Schemes             []string                    `json:"schemes"`                       // http/https
	Paths               map[string]map[string]Param `json:"paths"`                         // API 路径
	SecurityDefinitions *SecurityDefinitions        `json:"securityDefinitions,omitempty"` // 安全验证
	Definitions         map[string]Definition       `json:"definitions"`                   // 通用结构体定义
	ExternalDocs        *ExternalDocs               `json:"externalDocs,omitempty"`        // 外部链接
}
