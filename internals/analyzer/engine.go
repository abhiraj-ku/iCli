package analyzer

import (
	"encoding/json"
	"fmt"
)

func AnalyzePlan(planJSON string) ([]Issue, error) {
	var output ExplainOutput
	if err := json.Unmarshal([]byte(planJSON), &output); err != nil {
		return nil, fmt.Errorf("failed to parse EXPLAIN JSON: %w", err)
	}
	if len(output) == 0 {
		return nil, fmt.Errorf("empty EXPLAIN plan output")
	}

	var issues []Issue
	walk(&output[0].Plan, &issues)
	return issues, nil
}

func walk(node *PlanNode, issues *[]Issue) {
	if node == nil {
		return
	}
	*issues = append(*issues, evalNode(node)...)
	for _, child := range node.Plans {
		walk(child, issues)
	}
}
