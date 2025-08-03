package structs

type Command struct {
	Command    string  `json:"command"`
	Timestamp  string  `json:"timestamp"`
	AdjuntFile *string `json:"adjunt_file"`
}

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Type    string      `json:"type,omitempty"`
}