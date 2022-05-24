package utils

import "github.com/symphony09/running"

type StatesHelper struct {
	State running.State
}

func ProxyState(state running.State) StatesHelper {
	return StatesHelper{State: state}
}

func (helper StatesHelper) GetString(key string) (value string) {
	if helper.State == nil {
		return
	}

	raw, _ := helper.State.Query(key)
	value, _ = raw.(string)
	return
}

func (helper StatesHelper) GetInt(key string) (value int) {
	if helper.State == nil {
		return
	}

	raw, _ := helper.State.Query(key)
	value, _ = raw.(int)
	return
}

func (helper StatesHelper) GetFloat(key string) (value float64) {
	if helper.State == nil {
		return
	}

	raw, _ := helper.State.Query(key)
	value, _ = raw.(float64)
	return
}

func (helper StatesHelper) GetBool(key string) (value bool) {
	if helper.State == nil {
		return
	}

	raw, _ := helper.State.Query(key)
	value, _ = raw.(bool)
	return
}

func (helper StatesHelper) GetBytes(key string) (value []byte) {
	if helper.State == nil {
		return
	}

	raw, _ := helper.State.Query(key)
	value, _ = raw.([]byte)
	return
}
