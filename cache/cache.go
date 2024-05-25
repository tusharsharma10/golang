package cache

type Cache interface {
	SetJson(set string, key string, data interface{}, expiration int) error
	GetJson(set string, key string, container interface{}) (interface{}, error)
}
