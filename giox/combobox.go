package giox

// Combo holds combobox state
type Combo struct {
	items    []string
	selected int
}

// MakeCombo Creates new combobox widget
func MakeCombo(items []string) Combo {
	c := Combo{
		items:    items,
		selected: 0,
	}

	return c
}

// Selected returns currently selected item
func (c *Combo) Selected() string {
	return c.items[c.selected]
}
