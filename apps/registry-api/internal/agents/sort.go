package agents

import (
	"sort"
	"strings"
)

func sortAgents(items []Agent, sortBy, sortOrder string) {
	less := func(i, j int) bool {
		var cmp int
		switch sortBy {
		case "agent_id":
			cmp = strings.Compare(items[i].AgentID, items[j].AgentID)
		case "name":
			cmp = strings.Compare(items[i].Name, items[j].Name)
		case "updated_at":
			cmp = items[i].UpdatedAt.Compare(items[j].UpdatedAt)
		default:
			cmp = items[i].CreatedAt.Compare(items[j].CreatedAt)
		}
		if sortOrder == "asc" {
			return cmp < 0
		}
		return cmp > 0
	}
	sort.Slice(items, less)
}
