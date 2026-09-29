package analyzer

type ExplainOutput []struct {
	Plan PlanNode `json:"Plan"`
}

type PlanNode struct {
	NodeType   string      `json:"Node Type"`
	Relation   string      `json:"Relation Name,omitempty"`
	PlanRows   float64     `json:"Plan Rows"`
	TotalCost  float64     `json:"Total Cost"`
	Filter     string      `json:"Filter,omitempty"`
	SortMethod string      `json:"Sort Method,omitempty"`
	Plans      []*PlanNode `json:"Plans,omitempty"`
}

type Issue struct {
	Type        string
	Relation    string
	Description string
	Severity    string
}
