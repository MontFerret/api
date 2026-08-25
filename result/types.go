package result

type (
	Output struct {
		ContentType string `json:"contentType"`
		Content     []byte `json:"content"`
	}
)
