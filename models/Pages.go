package models

type Pages struct {
	Title   string
	Content string
	Data    interface{}
	Scripts []string
	Styles  []string
}

func (p *Pages) LoadDefaultScripts() {
	p.Scripts = []string{
		"https://unpkg.com/htmx.org@2.0.3",
		"https://cdn.tailwindcss.com",
	}
}
