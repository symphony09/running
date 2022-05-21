package utils

import "running"

type PropsHelper struct {
	Props running.Props
}

func ProxyProps(props running.Props) PropsHelper {
	return PropsHelper{Props: props}
}

func (helper PropsHelper) GetString(key string) (value string) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.Get(key)
	value, _ = raw.(string)
	return
}

func (helper PropsHelper) GetInt(key string) (value int) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.Get(key)
	value, _ = raw.(int)
	return
}

func (helper PropsHelper) GetFloat(key string) (value float64) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.Get(key)
	value, _ = raw.(float64)
	return
}

func (helper PropsHelper) GetBool(key string) (value bool) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.Get(key)
	value, _ = raw.(bool)
	return
}

func (helper PropsHelper) GetBytes(key string) (value []byte) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.Get(key)
	value, _ = raw.([]byte)
	return
}

func (helper PropsHelper) SubGetString(sub, key string) (value string) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.SubGet(sub, key)
	value, _ = raw.(string)
	return
}

func (helper PropsHelper) SubGetInt(sub, key string) (value int) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.SubGet(sub, key)
	value, _ = raw.(int)
	return
}

func (helper PropsHelper) SubGetFloat(sub, key string) (value float64) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.SubGet(sub, key)
	value, _ = raw.(float64)
	return
}

func (helper PropsHelper) SubGetBool(sub, key string) (value bool) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.SubGet(sub, key)
	value, _ = raw.(bool)
	return
}

func (helper PropsHelper) SubGetBytes(sub, key string) (value []byte) {
	if helper.Props == nil {
		return
	}

	raw, _ := helper.Props.SubGet(sub, key)
	value, _ = raw.([]byte)
	return
}
