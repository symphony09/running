package utils

import (
	"strconv"

	"github.com/symphony09/running"
)

type PropsHelper struct {
	SubKey string

	Props running.Props
}

func ProxyProps(props running.Props) PropsHelper {
	return PropsHelper{Props: props}
}

func (helper PropsHelper) Sub(subKey string) PropsHelper {
	helper.SubKey = subKey
	return helper
}

func (helper PropsHelper) GetRaw(key string) (value interface{}) {
	if helper.Props == nil {
		return
	}

	if helper.SubKey != "" {
		return helper.SubGetRaw(helper.SubKey, key)
	}

	raw, _ := helper.Props.Get(key)
	return raw
}

func (helper PropsHelper) GetString(key string) (value string) {
	if helper.Props == nil {
		return
	}

	if helper.SubKey != "" {
		return helper.SubGetString(helper.SubKey, key)
	}

	raw, _ := helper.Props.Get(key)
	return tranString(raw)
}

func (helper PropsHelper) GetInt(key string) (value int) {
	if helper.Props == nil {
		return
	}

	if helper.SubKey != "" {
		return helper.SubGetInt(helper.SubKey, key)
	}

	raw, _ := helper.Props.Get(key)
	return tranInt(raw)
}

func (helper PropsHelper) GetFloat(key string) (value float64) {
	if helper.Props == nil {
		return
	}

	if helper.SubKey != "" {
		return helper.SubGetFloat(helper.SubKey, key)
	}

	raw, _ := helper.Props.Get(key)
	return tranFloat(raw)
}

func (helper PropsHelper) GetBool(key string) (value bool) {
	if helper.Props == nil {
		return
	}

	if helper.SubKey != "" {
		return helper.SubGetBool(helper.SubKey, key)
	}

	raw, _ := helper.Props.Get(key)
	return tranBool(raw)
}

func (helper PropsHelper) GetBytes(key string) (value []byte) {
	if helper.Props == nil {
		return
	}

	if helper.SubKey != "" {
		return helper.SubGetBytes(helper.SubKey, key)
	}

	raw, _ := helper.Props.Get(key)
	value, _ = raw.([]byte)
	return
}

func (helper PropsHelper) SubGetRaw(sub, key string) (value interface{}) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.SubGet(sub, key)
	return raw
}

func (helper PropsHelper) SubGetString(sub, key string) (value string) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.SubGet(sub, key)
	return tranString(raw)
}

func (helper PropsHelper) SubGetInt(sub, key string) (value int) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.SubGet(sub, key)
	return tranInt(raw)
}

func (helper PropsHelper) SubGetFloat(sub, key string) (value float64) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.SubGet(sub, key)
	return tranFloat(raw)
}

func (helper PropsHelper) SubGetBool(sub, key string) (value bool) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.SubGet(sub, key)
	return tranBool(raw)
}

func (helper PropsHelper) SubGetBytes(sub, key string) (value []byte) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.SubGet(sub, key)
	value, _ = raw.([]byte)
	return
}

func tranString(raw interface{}) string {
	switch v := raw.(type) {
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	default:
		return ""
	}
}

func tranInt(raw interface{}) int {
	switch v := raw.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err != nil {
			return 0
		} else {
			return i
		}
	default:
		return 0
	}
}

func tranFloat(raw interface{}) float64 {
	switch v := raw.(type) {
	case int:
		return float64(v)
	case float64:
		return v
	case string:
		if f, err := strconv.ParseFloat(v, 64); err != nil {
			return 0
		} else {
			return f
		}
	default:
		return 0
	}
}

func tranBool(raw interface{}) bool {
	switch v := raw.(type) {
	case bool:
		return v
	case string:
		if b, err := strconv.ParseBool(v); err != nil {
			return false
		} else {
			return b
		}
	default:
		return false
	}
}
