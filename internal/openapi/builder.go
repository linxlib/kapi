package openapi

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/spec"
	"regexp"
	"strings"
)

type BaseBuilder struct {
	built bool
}

func (b *BaseBuilder) IsBuilt() bool {
	return b.built
}

type Builder struct {
	*BaseBuilder
	*spec.Swagger
}

func NewBuilder() *Builder {
	tmp := &Builder{
		BaseBuilder: &BaseBuilder{},
		Swagger: &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				ID:                  "",
				Consumes:            []string{"application/json"},
				Produces:            []string{"application/json"},
				Schemes:             []string{"http", "https"},
				Swagger:             "2.0",
				Info:                nil,
				Host:                "",
				BasePath:            "",
				Paths:               nil,
				Definitions:         nil,
				Parameters:          nil,
				Responses:           nil,
				SecurityDefinitions: nil,
				Security:            nil,
				Tags:                nil,
				ExternalDocs:        nil,
			},
		},
	}
	return tmp
}

func (b *Builder) String() string {
	bs, _ := json.Marshal(b)
	return string(bs)
}

func (b *Builder) WithInfo(docName string, docVer string, docDesc string) {
	b.Swagger.SwaggerProps.Info = &spec.Info{
		VendorExtensible: spec.VendorExtensible{
			Extensions: map[string]interface{}{
				"x-framework": "kapi",
				"x-version":   "v0.6.0",
			},
		},
		InfoProps: spec.InfoProps{
			Description: docDesc,
			Title:       docName,
			Contact: &spec.ContactInfo{

				ContactInfoProps: spec.ContactInfoProps{
					Name:  "kapi",
					URL:   "https://github.com/linxlib/kapi",
					Email: "",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "MIT",
					URL:  "https://github.com/linxlib/kapi/blob/main/LICENSE",
				},
			},
			Version: docVer,
		},
	}
}

func (b *Builder) SetHost(host string) {
	b.Swagger.SwaggerProps.Host = host
}

func (b *Builder) AddTag(tag spec.Tag) {
	if b.Swagger.Tags == nil {
		b.Swagger.Tags = []spec.Tag{}
	}
	for _, t := range b.Swagger.Tags {
		if t.Name == tag.Name {
			return
		}
	}
	b.Swagger.Tags = append(b.Swagger.Tags, tag)
}
func (b *Builder) AddDefinitions(name string, definitions spec.Schema) {
	if b.Swagger.Definitions == nil {
		b.Swagger.Definitions = make(spec.Definitions)
	}

	b.Swagger.Definitions[name] = definitions
}

var pathRouterRegex = regexp.MustCompile(":\\w+")

// replacePathTo 将:id转为{id}这样的格式
func replacePathTo(origin string) string {
	var myorigin = origin
	matches := pathRouterRegex.FindAllString(origin, -1)

	if len(matches) > 0 {
		for _, match := range matches {
			if strings.HasPrefix(match, ":") {
				pathName := strings.TrimPrefix(match, ":")
				pathName = strings.TrimSuffix(pathName, "/")
				pathName = fmt.Sprintf("{%s}", pathName)
				myorigin = strings.ReplaceAll(myorigin, match, pathName)
			}
		}
		return myorigin
	}
	return myorigin
}

func (b *Builder) AddRoute(method string, path string, p *spec.Operation) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = replacePathTo(path)
	if b.Swagger.Paths == nil || b.Swagger.Paths.Paths == nil {
		b.Swagger.Paths = &spec.Paths{
			Paths: make(map[string]spec.PathItem),
		}
	}

	switch method {
	case "GET":

		b.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Get: p,
			},
		}
	case "POST":
		b.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: p,
			},
		}
	case "PUT":
		b.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Put: p,
			},
		}
	case "DELETE":
		b.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Delete: p,
			},
		}
	case "OPTIONS":
		b.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Options: p,
			},
		}
	case "HEAD":
		b.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Head: p,
			},
		}
	case "PATCH":
		b.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Patch: p,
			},
		}
	}

}

func (b *Builder) Build() *spec.Swagger {
	b.built = true
	return b.Swagger
}

type TagBuilder struct {
	*BaseBuilder
	spec.Tag
}

func NewTagBuilder() TagBuilder {
	return TagBuilder{BaseBuilder: &BaseBuilder{}, Tag: spec.Tag{}}
}
func (t *TagBuilder) Build() spec.Tag {
	t.built = true
	return t.Tag
}

type RefBuilder struct {
	*BaseBuilder
	spec.Ref
}

func NewRefBuilder(ref string) *RefBuilder {
	return &RefBuilder{
		BaseBuilder: &BaseBuilder{},
		Ref:         spec.MustCreateRef(ref),
	}
}
func (r *RefBuilder) Build() spec.Ref {
	r.built = true
	return r.Ref
}

type SchemaBuilder struct {
	*BaseBuilder
	spec.Schema
}

func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{BaseBuilder: &BaseBuilder{}, Schema: spec.Schema{}}
}
func (s *SchemaBuilder) Build() spec.Schema {
	s.built = true
	return s.Schema
}

type OperationBuilder struct {
	*BaseBuilder
	*spec.Operation
}

func NewOperationBuilder() *OperationBuilder {
	return &OperationBuilder{BaseBuilder: &BaseBuilder{}, Operation: &spec.Operation{
		OperationProps: spec.OperationProps{},
	}}
}

func (o *OperationBuilder) Build() *spec.Operation {
	o.built = true
	return o.Operation
}

type ParameterBuilder struct {
	*BaseBuilder
	*spec.Parameter
}

func NewParameterBuilder() *ParameterBuilder {
	return &ParameterBuilder{
		BaseBuilder: &BaseBuilder{},
		Parameter:   &spec.Parameter{},
	}
}

func (p *ParameterBuilder) Build() *spec.Parameter {
	p.built = true
	return p.Parameter
}

type ResponseBuilder struct {
	*BaseBuilder
	*spec.Response
}

func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		BaseBuilder: &BaseBuilder{},
		Response: &spec.Response{
			ResponseProps: spec.ResponseProps{},
		},
	}
}

func (r *ResponseBuilder) Build() *spec.Response {
	r.built = true
	return r.Response
}
