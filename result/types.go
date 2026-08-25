package result

type (
	Output struct {
		// ContentType is the MIME type of the encoded content.
		ContentType string
		// Content holds the encoded output bytes.
		Content []byte
	}
)
