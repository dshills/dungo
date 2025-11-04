package dungeon

import "github.com/dshills/dungo/pkg/graph"

// Graph is the Abstract Dungeon Graph container.
// It wraps the underlying graph.Graph with additional metadata.
type Graph struct {
	*graph.Graph // Embedded graph with all operations
}
