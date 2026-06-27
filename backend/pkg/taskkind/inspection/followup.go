package inspection

// checkItemConfig and branchConfig are kept for backward-compatible migration
// from old JSON CheckItems format. They are not used in the new table-based model.

type branchConfig struct {
	Name       string `json:"name"`
	CreateTodo bool   `json:"create_todo"`
	TodoName   string `json:"todo_name"`
	GroupID    string `json:"group_id"`
}

type checkItemConfig struct {
	Name     string         `json:"name"`
	Branches []branchConfig `json:"branches"`
}
