package dummy

type DoesTheThings interface {
	Get(key string) error
	Put(key string, val string) error
}
