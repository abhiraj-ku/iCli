package analyzer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/abhiraj-ku/pg_adv/internals/db"
)

type N1Issue struct {
	QueryID         int64
	QueryText       string
	Calls           int64
	TotalTime       float64
	AvgTime         float64
	AvgRowsPerCall  float64
	PatternType     string
	TargetTable     string
	FilterColumn    string
	Severity        string
	ParentCandidate *db.N1Stat
	Recommendation  string
}

type N1ReportItem = N1Issue

var (
	paramRe = regexp.MustCompile(`(?i)\bWHERE\s+([a-zA-Z0-9_"\.]+)\s*=\s*(\$1|\?)`)
	fromRe  = regexp.MustCompile(`(?i)\bFROM\s+([a-zA-Z0-9_"\.]+)`)
)

func DetectN1(stats []db.N1Stat) []N1Issue {
	var out []N1Issue

	for i, stat := range stats {
		if !isSelect(stat.QueryText) {
			continue
		}

		avgRows := 0.0
		if stat.Calls > 0 {
			avgRows = float64(stat.Rows) / float64(stat.Calls)
		}

		col, isParam := parseFilter(stat.QueryText)

		if isParam || (avgRows > 0 && avgRows <= 10.0 && stat.Calls >= 20) {
			table := parseTable(stat.QueryText)
			sev := scoreSeverity(stat.Calls, stat.TotalTime)

			rec := N1Issue{
				QueryID:        stat.QueryID,
				QueryText:      stat.QueryText,
				Calls:          stat.Calls,
				TotalTime:      stat.TotalTime,
				AvgTime:        stat.AvgTime,
				AvgRowsPerCall: avgRows,
				TargetTable:    table,
				FilterColumn:   col,
				Severity:       sev,
			}

			if isParam {
				rec.PatternType = fmt.Sprintf("Parametrized Single-Row Lookup (`WHERE %s = $1`)", col)
			} else {
				rec.PatternType = "High-Frequency Iterative Fetch (Low Row Yield per Call)"
			}

			if parent := matchParent(stat, stats, i); parent != nil {
				rec.ParentCandidate = parent
			}

			rec.Recommendation = recommendation(rec)
			out = append(out, rec)
		}
	}

	return out
}

func DetectN1Queries(stats []db.N1Stat) []N1Issue {
	return DetectN1(stats)
}

func isSelect(q string) bool {
	s := strings.TrimSpace(strings.ToUpper(q))
	return strings.HasPrefix(s, "SELECT") || strings.HasPrefix(s, "WITH")
}

func parseFilter(q string) (string, bool) {
	m := paramRe.FindStringSubmatch(q)
	if len(m) >= 2 {
		return m[1], true
	}
	return "", false
}

func parseTable(q string) string {
	m := fromRe.FindStringSubmatch(q)
	if len(m) >= 2 {
		return m[1]
	}
	return "unknown"
}

func scoreSeverity(calls int64, total float64) string {
	if calls >= 500 || total >= 1000.0 {
		return "High"
	}
	if calls >= 100 || total >= 200.0 {
		return "Medium"
	}
	return "Low"
}

func matchParent(child db.N1Stat, list []db.N1Stat, idx int) *db.N1Stat {
	for i, parent := range list {
		if i == idx || !isSelect(parent.QueryText) {
			continue
		}

		if parent.Calls < child.Calls {
			pRows := parent.Rows
			if pRows > 0 {
				ratio := float64(child.Calls) / float64(pRows)
				if ratio >= 0.5 && ratio <= 2.0 {
					return &list[i]
				}
			}
			if parent.Calls > 0 && (child.Calls/parent.Calls) >= 5 {
				return &list[i]
			}
		}
	}
	return nil
}

func recommendation(rec N1Issue) string {
	var b strings.Builder
	if rec.FilterColumn != "" {
		b.WriteString(fmt.Sprintf("Replace repeated single lookups (`WHERE %s = $1`) with batched array fetching:\n", rec.FilterColumn))
		b.WriteString(fmt.Sprintf("  SELECT * FROM %s WHERE %s = ANY($1::int[]);\n", rec.TargetTable, rec.FilterColumn))
		b.WriteString("Or configure Eager Loading (JOIN / prefetch) in your ORM (e.g. GORM `.Preload()`, Prisma `include`, TypeORM `relations`).")
	} else {
		b.WriteString(fmt.Sprintf("Batch queries on table `%s` using `WHERE ... IN (...)` or combine queries using JOINs/subqueries instead of running queries in application loops.", rec.TargetTable))
	}
	return b.String()
}
