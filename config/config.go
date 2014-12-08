package config

type Config struct {

	PersistInMemory      bool
	ObjectIdentifier     string
	CollectionIdentifier string
}

var DefaultConfig = Config{false, "objectId", "results"}
