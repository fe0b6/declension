package declension

type gender struct {
	Gender map[string]map[string][]string `json:"gender"`
}

type ruleData struct {
	Gender string   `json:"gender"`
	Test   []string `json:"test"`
	Mods   []string `json:"mods"`
	Tags   []string `json:"tags"`
}
