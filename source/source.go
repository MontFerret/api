package source

type Source struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func New(name, content string) Source {
	return Source{
		Name:    name,
		Content: content,
	}
}

func NewAnonymous(content string) Source {
	return New("anonymous", content)
}
