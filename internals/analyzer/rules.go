package analyzer

import "fmt"

func evalNode(node *PlanNode) []Issue {
	var issues []Issue

	if issue, found := checkSeqScan(node); found {
		issues = append(issues, issue)
	}
	if issue, found := checkDiskSort(node); found {
		issues = append(issues, issue)
	}
	if issue, found := checkNestedLoop(node); found {
		issues = append(issues, issue)
	}
	return issues
}

func checkSeqScan(node *PlanNode) (Issue, bool) {
	if node.NodeType == "Seq Scan" && node.Filter != "" && node.TotalCost > 50 {
		return Issue{
			Type:        "Missing Index",
			Relation:    node.Relation,
			Description: fmt.Sprintf("Heavy Sequential Scan detected (Cost: %.2f). Consider an index on the filtered columns.", node.TotalCost),
			Severity:    "High",
		}, true
	}
	return Issue{}, false
}

func checkDiskSort(node *PlanNode) (Issue, bool) {
	if node.NodeType == "Sort" && node.SortMethod == "external merge disk" {
		return Issue{
			Type:        "Memory Starvation",
			Relation:    "N/A",
			Description: "Sort operation spilled to disk. Increase 'work_mem' or add an index to pre-sort the data.",
			Severity:    "High",
		}, true
	}
	return Issue{}, false
}

func checkNestedLoop(node *PlanNode) (Issue, bool) {
	if node.NodeType == "Nested Loop" {
		for _, child := range node.Plans {
			if child.NodeType == "Seq Scan" && child.PlanRows > 1000 {
				return Issue{
					Type:        "Inefficient Join",
					Relation:    child.Relation,
					Description: "Nested Loop driving a heavy Sequential Scan. This causes massive I/O amplification. Index the join condition.",
					Severity:    "High",
				}, true
			}
		}
	}
	return Issue{}, false
}
