package analyzer

import (
	"encoding/json"
	"fmt"
)

// parses the raw JSON execution plan and returns a list of identified issues.
func AnalyzePlan(planJson string) ([]Issue, error) {
	var output ExplainOutput

	// parse the raw json into our strutc
	if err := json.Unmarshal([]byte(planJson), &output); err != nil {
		return nil, fmt.Errorf("failed tp parse Explain Json(): %w", err)
	}
	if len(output) == 0 {
		return nil, fmt.Errorf("empty explain plan found")
	}
	var issue []Issue

	// inti the reverse tree walk algo
	walk(&output[0].Plan, &issue)

	return issue, nil
}

// Recirsive tree walk and applying heuristic at each node s
func walk(node *PlanNode, issues *[]Issue) {
	// base case
	if node == nil {
		return
	}
	// eval the current node gainsta all rule
	nodeIssues := evalNode(node)
	*issues = append(*issues, nodeIssues...)

	// traverse all choldrem node

	for _, child := range node.Plans {
		walk(child, issues)
	}
}
