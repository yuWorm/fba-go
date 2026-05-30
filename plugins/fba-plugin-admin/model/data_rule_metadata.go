package model

type DataRuleColumn struct {
	Key     string
	Comment string
}

type DataRuleTemplateVariable struct {
	Key     string
	Comment string
}

type DataRuleModelMetadata struct {
	Name    string
	Columns []DataRuleColumn
}
