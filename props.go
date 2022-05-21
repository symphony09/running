package running

type StandardProps map[string]interface{}

func (props StandardProps) Get(key string) (value interface{}, exists bool) {
	value, exists = props[key]
	return
}

func (props StandardProps) SubGet(sub, key string) (value interface{}, exists bool) {
	return props.Get(sub + "." + key)
}

type EmptyProps struct{}

func (props EmptyProps) Get(key string) (value interface{}, exists bool) {
	return
}

func (props EmptyProps) SubGet(sub, key string) (value interface{}, exists bool) {
	return
}
