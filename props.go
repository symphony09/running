package running

type StandardProps map[string]interface{}

func (props StandardProps) Get(key string) (value interface{}, exists bool) {
	value, exists = props[key]
	return
}

type EmptyProps struct{}

func (props EmptyProps) Get(key string) (value interface{}, exists bool) {
	return
}
