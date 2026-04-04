package zk

type Adapter interface {
	Health() ZKStatus
	IsReadOnly() ZKStatus
	Ping() ZKStatus
	Close() error
}

// Personal.AI order the ending
