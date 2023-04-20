package utils

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/symphony09/running"
)

// RegisterNodes auto register node builder, field with running tag will be set
// tag `running:"name"` to get node name
// tag `running:"prop:key"` to get prop value of the key
func RegisterNodes(e *running.Engine, nodes ...running.Node) error {
	for _, node := range nodes {
		name, builder, props, err := parseNode(node)
		if err != nil {
			return err
		} else {
			e.RegisterNodeBuilder(name, builder)

			var info running.NodeBuilderInfo
			if _, ok := node.(running.Cluster); ok {
				info.Type = running.TypeOfCluster
			} else if _, ok = node.(running.Wrapper); ok {
				info.Type = running.TypeOfWrapper
			} else {
				info.Type = running.TypeOfCommon
			}

			if pc, _, _, ok := runtime.Caller(1); ok {
				info.From = runtime.FuncForPC(pc).Name()
			}

			info.Note = fmt.Sprintf("Property Map: %+v\nAuto regitered by util of runnning.", props)

			e.SetNodeBuilderInfo(name, info)
		}
	}

	return nil
}

// RegisterNodeWithTypeName similar to RegisterNodes, but specify the type name of node
func RegisterNodeWithTypeName(e *running.Engine, typeName string, node running.Node) error {
	_, builder, props, err := parseNode(node)
	if err != nil {
		return err
	} else {
		e.RegisterNodeBuilder(typeName, builder)

		var info running.NodeBuilderInfo
		if _, ok := node.(running.Cluster); ok {
			info.Type = running.TypeOfCluster
		} else if _, ok = node.(running.Wrapper); ok {
			info.Type = running.TypeOfWrapper
		} else {
			info.Type = running.TypeOfCommon
		}

		if pc, _, _, ok := runtime.Caller(1); ok {
			info.From = runtime.FuncForPC(pc).Name()
		}

		info.Note = fmt.Sprintf("Propery Map: %+v \nAuto regitered by util of runnning.", props)

		e.SetNodeBuilderInfo(typeName, info)

		return nil
	}
}

type Uninitialized interface {
	Init() error
}

func parseNode(node running.Node) (typeName string, builder running.BuildNodeFunc, props map[string]reflect.Type, err error) {
	nodeType := reflect.TypeOf(node)
	if nodeType.Kind() == reflect.Ptr {
		nodeType = nodeType.Elem()
	}

	if nodeType.Kind() != reflect.Struct {
		err = fmt.Errorf("non struct type kind not supported, got node type kind = %v", nodeType.Kind())
		return
	}

	props = make(map[string]reflect.Type)

	typeName = strings.TrimPrefix(nodeType.Name(), nodeType.PkgPath())
	autowired := map[string]string{}
	var nameField, baseField string

	for i := 0; i < nodeType.NumField(); i++ {
		f := nodeType.Field(i)
		tag, ok := f.Tag.Lookup("running")
		if ok {
			tagItems := strings.Split(tag, ";")
			for _, item := range tagItems {
				k, v, found := strings.Cut(item, ":")
				if found {
					switch strings.TrimSpace(k) {
					case "prop":
						propName := strings.TrimSpace(v)
						autowired[f.Name] = propName
						props[propName] = f.Type
					}
				} else {
					switch strings.TrimSpace(k) {
					case "name":
						nameField = f.Name
					}
				}
			}
		} else if f.Anonymous && f.Type == reflect.TypeOf(running.Base{}) {
			baseField = f.Name
		}

		nodeVal := reflect.New(nodeType)
		for fieldName := range autowired {
			field := nodeVal.Elem().FieldByName(fieldName)

			if !field.CanSet() {
				err = fmt.Errorf("props field %s cannot be set", fieldName)
				return
			}
		}

		if nameField != "" {
			field := nodeVal.Elem().FieldByName(nameField)

			if field.Type().String() != "string" {
				err = fmt.Errorf("name field %s must be string type", nameField)
				return
			}

			if !field.CanSet() {
				err = fmt.Errorf("name field %s cannot be set", nameField)
				return
			}
		}
	}

	builder = func(name string, props running.Props) (newNode running.Node, err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("cannot build node, %v", e)
			}
		}()

		newNodeVal := reflect.New(nodeType)
		newNode = newNodeVal.Interface().(running.Node)

		for fieldName, propName := range autowired {
			val, found := props.SubGet(name, propName)
			if !found {
				continue
			}

			field := newNodeVal.Elem().FieldByName(fieldName)

			if field.Type() != reflect.TypeOf(val) {
				err = fmt.Errorf("prop field type = %v, prop value type = %v", field.Type(), reflect.TypeOf(val))
				return
			}

			field.Set(reflect.ValueOf(val))
		}

		if nameField != "" {
			newNodeVal.Elem().FieldByName(nameField).SetString(name)
		}

		if baseField != "" {
			newNodeVal.Elem().FieldByName(baseField).FieldByName("NodeName").SetString(name)
		}

		if uninitializedNode, ok := newNode.(Uninitialized); ok {
			err = uninitializedNode.Init()
		}

		return
	}

	return
}
