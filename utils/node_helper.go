package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/symphony09/running"
)

// RegisterNodes auto register node builder, field with running tag will be set
// tag `running:"name"` to get node name
// tag `running:"prop:key"` to get prop value of the key
func RegisterNodes(e *running.Engine, nodes ...running.Node) error {
	for _, node := range nodes {
		name, builder, err := parseNode(node)
		if err != nil {
			return err
		} else {
			e.RegisterNodeBuilder(name, builder)
		}
	}

	return nil
}

type Uninitialized interface {
	Init() error
}

func parseNode(node running.Node) (typeName string, builder running.BuildNodeFunc, err error) {
	nodeType := reflect.TypeOf(node)
	if nodeType.Kind() == reflect.Ptr {
		nodeType = nodeType.Elem()
	}

	if nodeType.Kind() != reflect.Struct {
		err = fmt.Errorf("non struct type kind not supported, got node type kind = %v", nodeType.Kind())
		return
	}

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
						autowired[f.Name] = strings.TrimSpace(v)
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
