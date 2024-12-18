package store

type NamedKey interface {
	GetNamedKey() string
}

type Key interface {
	NamedKey

	Marshal() ([]byte, error)

	UnMarshal(key string, data []byte) error
}

type KeyCreator func() Key

type Observer interface {
	OnUpdate(key Key)
	OnDelete(key NamedKey)
}
