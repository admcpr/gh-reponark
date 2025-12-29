package shared

type ListItem struct {
	title, desc string
}

func NewListItem(title, desc string) ListItem {
	return ListItem{title, desc}
}

func (i ListItem) Title() string       { return i.title }
func (i ListItem) Description() string { return i.desc }
func (i ListItem) FilterValue() string { return i.title }
