package config

type Config struct {

	/**
	 * Data source response configurations
	 */
	// endpoint url of the http server
	HTTPServerURI                   string `json:"serverURL,omitempty"`

	// key value of the objects that comes from data source (ex: 'id', '_id', 'objectId')
	ObjectIdentifier                string `json:"objectIdentifier,omitempty"`

	// key value of the lists that comes from data source (ex: 'list', 'results').
	// leave it empty if array comes in the root of the data received from data source
	CollectionIdentifier            string `json:"collectionIdentifier,omitempty"`


	/**
	 * Server behaviour configurations
	 */
	// keeps data in memory and responds back on 'get' requests without asking to data source
	PersistItemInMemory             bool `json:"persistItemInMemory"`

	// keeps lists references in memory and works same as PersistItemInMemory
	PersistListInMemory             bool `json:"persistListInMemory"`

	// destroys the nodes when all subscribers(ws-connections) are gone
	CleanupOnSubscriptionsOver      bool `json:"cleanupOnSubscriptionsOver"`
}

var SystemConfig Config
