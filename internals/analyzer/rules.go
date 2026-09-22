package analyzer

import "fmt"

// Runs all heuristic rules againsts a single exex plan node
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

// identifies large table scans with filter
func checkSeqScan(node *PlanNode) (Issue, bool) {
	if node.NodeType == "Seq Scan" && node.Filter != "" && node.PlanRows > 1000 {
		return Issue{
			Type:        "Missing Index",
			Relation:    node.Relation,
			Description: fmt.Sprintf("Sequential scan filtering %.0f projected rows. Consider an index on the filtered columns.", node.PlanRows),
			Severity:    "High",
		}, true
	}

	return Issue{}, false
}

// identifies sorts that exhausted work_mem and spilled to the hard drive.
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

// identifies devastating N+1 queries driven by an outer loop.
func checkNestedLoop(node *PlanNode) (Issue, bool) {
	if node.NodeType == "Nested Loop" {
		// inspect and see if the childrens are doing full page scan
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
