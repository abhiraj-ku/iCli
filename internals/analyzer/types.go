package analyzer

// represents the root array returned by EXPLAIN (FORMAT JSON).
type ExplainOutput []struct {
	Plan PlanNode `json:"Plan"`
}

// recursive AST node representing a step in the query execution.
type PlanNode struct {
	NodeType   string      `json:"Node Type"`
	Relation   string      `json:"Relation Name,omitempty"`
	PlanRows   float64     `json:"Plan Rows"`
	Filter     string      `json:"Filter,omitempty"`
	SortMethod string      `json:"Sort Method,omitempty"`
	Plans      []*PlanNode `json:"Plans,omitempty"`
}

// represents a performance bottleneck identified by the heuristics engine.
type Issue struct {
	Type        string
	Relation    string
	Description string
	Severity    string // High, Medium, Low
}
