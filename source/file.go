package source

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func New(name, content string) File {
	return File{
		Name:    name,
		Content: content,
	}
}

func NewAnonymous(content string) File {
	return New("anonymous", content)
}
