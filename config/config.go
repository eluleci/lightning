package config

type Config struct {

	/**
	 * Data source response configurations
	 */
	// the key value of the objects that comes from data source (ex: 'id', '_id', 'objectId')
	ObjectIdentifier                string

	// the key value of the lists that comes from data source (ex: 'list', 'results').
	// leave it empty if array comes in the root of the data received from data source
	CollectionIdentifier            string


	/**
	 * Server behaviour configurations
	 */
	// keeps data in memory and responds back on 'get' requests without asking to data source
	PersistItemInMemory             bool

	// keeps lists references in memory and works same as PersistItemInMemory
	PersistListInMemory             bool

	// destroys the nodes when there is all subscribers(ws-connections) are gone
	CleanupOnSubscriptionsOver      bool
}

var DefaultConfig = Config{"objectId", "results", true, true, false}
