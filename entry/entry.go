package entry

import "time"

type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	// TODO consider using a more allocation-efficient representation
	Record map[string]interface{} `json:"record"`
}

func CreateBasicEntry(message interface{}) Entry {
	return Entry{
		Timestamp: time.Now(),
		Record: map[string]interface{}{
			"message": message,
		},
	}
}
