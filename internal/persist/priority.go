package persist

type PriorityPersister struct {
	pers     Persister
	name     string
	priority int
}

func NewPriorityPersister(pers Persister) *PriorityPersister {
	var (
		prior int
		name  string
	)

	switch pers.(type) {
	case *RedisPersister:
		prior = 1
		name = "redis"
	case *PostgresPersister:
		prior = 2
		name = "postgres"
	case *FilePersister:
		prior = 3
		name = "file"
	}

	return &PriorityPersister{
		pers:     pers,
		priority: prior,
		name:     name,
	}
}
