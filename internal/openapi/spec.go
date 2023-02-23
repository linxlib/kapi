package openapi

import (
	"github.com/go-openapi/spec"
	"github.com/linxlib/kapi/internal"
	"github.com/linxlib/kapi/internal/ast_parser"
	"strings"
)

type Spec struct {
	*Builder
}

func NewSpec() *Spec {
	return &Spec{
		NewBuilder(),
	}
}
func (myspec *Spec) AddRoute(method string, path string, deprecated bool, summary string, tag string, requestParams []*spec.Parameter, responseParams []*spec.Response) {
	op := spec.NewOperation(method + path)
	op.Deprecated = deprecated
	op.WithSummary(summary).WithTags(tag)
	for _, param := range requestParams {
		op.AddParam(param)
	}
	for _, param := range responseParams {
		op.RespondsWith(200, param)
	}

	myspec.normalResponse(op)

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = replacePathTo(path)
	if myspec.Swagger.Paths == nil || myspec.Swagger.Paths.Paths == nil {
		myspec.Swagger.Paths = &spec.Paths{
			Paths: make(map[string]spec.PathItem),
		}
	}

	switch method {
	case "GET":

		myspec.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Get: op,
			},
		}
	case "POST":
		myspec.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: op,
			},
		}
	case "PUT":
		myspec.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Put: op,
			},
		}
	case "DELETE":
		myspec.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Delete: op,
			},
		}
	case "OPTIONS":
		myspec.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Options: op,
			},
		}
	case "HEAD":
		myspec.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Head: op,
			},
		}
	case "PATCH":
		myspec.Swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Patch: op,
			},
		}
	}
}

func (myspec *Spec) normalResponse(opBuilder *spec.Operation) {
	schema := &spec.Schema{}
	schema.WithDescription("parameter error or validation failed\")")
	p := &spec.Schema{}
	p.WithDescription("code").
		Typed("integer", "").
		WithDefault(0)
	schema.SetProperty("code", *p)

	p1 := &spec.Schema{}
	p1.WithDescription("message").
		Typed("string", "")
	schema.SetProperty("msg", *p1)
	p2 := &spec.Schema{}
	p2.WithDescription("data").
		Typed("object", "")
	schema.SetProperty("data", *p2)
	opBuilder.RespondsWith(400, spec.NewResponse().
		WithDescription("parameter error or validation failed").
		WithSchema(schema),
	)
}

func (myspec *Spec) ResponseParams(sResp *ast_parser.Struct) (responseParams []*spec.Response) {
	responseParams = make([]*spec.Response, 0)

	if sResp != nil {
		myspec.definitionSchema(sResp)

		response := spec.NewResponse().
			WithDescription(strings.Join(sResp.Docs, "\n")).
			WithSchema(spec.RefSchema("#/definitions/" + sResp.Name))
		responseParams = append(responseParams, response)
	}
	return responseParams
}

func (myspec *Spec) RequestParams(sReq *ast_parser.Struct) (requestParams []*spec.Parameter) {
	requestParams = make([]*spec.Parameter, 0)
	if sReq != nil {
		// build json body scheme and add it into definitions
		myspec.definitionSchema(sReq)
		if len(sReq.GetAllFieldsByTag("json")) > 0 {
			requestParams = append(requestParams,
				spec.BodyParam(sReq.Name,
					spec.RefSchema("#/definitions/"+sReq.Name)).
					WithDescription(strings.Join(sReq.Docs, "\n")).
					AsRequired())
		}

		params := myspec.Parameter(sReq, "query")
		if len(params) > 0 {
			requestParams = append(requestParams, params...)
		}
		params = myspec.Parameter(sReq, "header")
		if len(params) > 0 {
			requestParams = append(requestParams, params...)
		}
		params = myspec.Parameter(sReq, "form")
		if len(params) > 0 {
			requestParams = append(requestParams, params...)
		}
		params = myspec.Parameter(sReq, "path")
		if len(params) > 0 {
			requestParams = append(requestParams, params...)
		}

		// TODO: 使用form时 POST请求会解析失败

	}
	return requestParams
}

func (myspec *Spec) Parameter(sReq *ast_parser.Struct, tag string) (params []*spec.Parameter) {
	params = make([]*spec.Parameter, 0)
	queryFields := sReq.GetAllFieldsByTag(tag)

	makeParam := func(tag string, field *ast_parser.Field) *spec.Parameter {
		switch tag {
		case "query":
			return spec.QueryParam(field.CurrentTag)
		case "header":
			return spec.HeaderParam(field.CurrentTag)
		case "path":
			return spec.PathParam(field.CurrentTag)
		case "form":
			return spec.FormDataParam(field.CurrentTag)
		case "file":
			return spec.FileParam(field.CurrentTag)
		default:
			panic("not supported tag")
		}
	}

	for _, field := range queryFields {
		parameter := makeParam(tag, field).
			WithDescription(field.Comment)
		if field.GetTag("v") == "required" {
			parameter.AsRequired()
		}
		if field.GetTag("default") != "" {
			parameter.WithDefault(field.GetTag("default"))
		}
		//TODO: get param type
		if strings.HasPrefix(field.Type, "[]") {
			parameter.Typed("array", "")
			if field.IsStruct {
				if !field.Struct.IsEnum {
					panic("struct parameter is not supported")
					// normal param should not be struct model
				} else {
					parameter.Enum = make([]interface{}, 0)
					for _, f := range field.Struct.Fields {
						parameter.Enum = append(parameter.Enum, f.EnumValue)
					}
				}
			} else {
				parameter.Items = &spec.Items{}
				parameter.Items.Type = internal.GetType(strings.TrimPrefix(field.Type, "[]"))
			}

		} else {
			if field.IsStruct {
				if !field.Struct.IsEnum {
					panic("struct parameter is not supported")
					// query param should not be struct model
				} else {
					//parameter.Typed(field.Struct.Name, "")
					parameter.Enum = make([]interface{}, 0)
					for _, f := range field.Struct.Fields {
						parameter.Enum = append(parameter.Enum, f.EnumValue)
					}
				}
			} else {
				parameter.Typed(internal.GetType(field.Type), internal.GetFormat(field.Type, ""))
			}

		}
		params = append(params, parameter)
	}
	return params
}

func (myspec *Spec) definitionSchema(s *ast_parser.Struct) {
	fds := s.GetAllFieldsByTag("json")
	bodyDefineSchema := spec.Schema{}
	bodyDefineSchema.WithDescription(strings.Join(s.Docs, "\n"))
	if s.IsEnum {
		//enum fields can not be struct, must be inner type
		bodyDefineSchema.Typed(internal.GetType(s.EnumType), "")
		bodyDefineSchema.Enum = make([]interface{}, 0)
		for _, field := range fds { //iter enum fields
			bodyDefineSchema.Enum = append(bodyDefineSchema.Enum, field.EnumValue)
		}
		myspec.AddDefinitions(s.Name, bodyDefineSchema)
	} else {
		bodyDefineSchema.Typed("object", "") //request body or response body must be struct. should not be array

		for _, field := range fds { //iter struct fields
			fieldName := field.CurrentTag

			bodyFieldSchema := spec.Schema{}
			bodyFieldSchema.WithDescription(field.Comment)
			if field.GetTag("default") != "" {
				bodyFieldSchema.WithDefault(field.GetTag("default"))
			}
			if strings.Contains(field.GetTag("v"), "required") {
				bodyDefineSchema.AddRequired(fieldName)
			}
			if strings.HasPrefix(field.Type, "[]") { //field can be slice
				bodyFieldSchema.Typed("array", "")
				if field.IsStruct {
					if !field.Struct.IsEnum {
						myspec.definitionSchema(field.Struct)
						bodyFieldSchema.Ref = spec.MustCreateRef("#/definitions/" + field.Struct.Name)
					} else {
						bodyFieldSchema.Enum = make([]interface{}, 0)
						for _, f := range field.Struct.Fields {
							bodyFieldSchema.Enum = append(bodyFieldSchema.Enum, f.EnumValue)
						}
					}
				} else {
					//add an item schema
					ss := NewSchemaBuilder()
					ss.AddType(internal.GetType(strings.TrimPrefix(field.Type, "[]")), internal.GetFormat(field.Type, ""))
					ss.Build()
					bodyFieldSchema.Items = &spec.SchemaOrArray{}
					bodyFieldSchema.Items.Schema = &ss.Schema
				}

			} else {
				if field.IsStruct {
					//enum type not handled here
					if !field.Struct.IsEnum {
						myspec.definitionSchema(field.Struct)
						bodyFieldSchema.Ref = spec.MustCreateRef("#/definitions/" + field.Struct.Name)
					} else {
						bodyFieldSchema.Enum = make([]interface{}, 0)
						for _, f := range field.Struct.Fields {
							bodyFieldSchema.Enum = append(bodyFieldSchema.Enum, f.EnumValue)
						}
					}

				} else {
					bodyFieldSchema.Typed(internal.GetType(field.Type), internal.GetFormat(field.Type, ""))
				}

			}
			bodyDefineSchema.SetProperty(fieldName, bodyFieldSchema)

		}
	}

	myspec.AddDefinitions(s.Name, bodyDefineSchema)
}
