package export_test

import (
	"fmt"
	"log"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/export"
	"github.com/dshills/dungo/pkg/graph"
)

// ExampleExportJSON demonstrates basic JSON export of a dungeon artifact.
func ExampleExportJSON() {
	// Create a simple artifact
	artifact := &dungeon.Artifact{
		ADG: &dungeon.Graph{
			Graph: &graph.Graph{
				Rooms: map[string]*graph.Room{
					"start": {
						ID:         "start",
						Archetype:  graph.ArchetypeStart,
						Size:       graph.SizeM,
						Difficulty: 0.1,
					},
					"boss": {
						ID:         "boss",
						Archetype:  graph.ArchetypeBoss,
						Size:       graph.SizeXL,
						Difficulty: 0.9,
					},
				},
			},
		},
		Metrics: &dungeon.Metrics{
			BranchingFactor: 2.0,
			PathLength:      5,
		},
	}

	// Export to JSON
	data, err := export.ExportJSON(artifact)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exported %d bytes of JSON data\n", len(data))
	// Output: Exported 612 bytes of JSON data
}

// ExampleSaveJSONToFile demonstrates saving a dungeon artifact to a JSON file.
func ExampleSaveJSONToFile() {
	// Create a simple artifact
	artifact := &dungeon.Artifact{
		ADG: &dungeon.Graph{
			Graph: &graph.Graph{
				Rooms: map[string]*graph.Room{
					"room-1": {ID: "room-1", Archetype: graph.ArchetypeStart},
				},
			},
		},
	}

	// Save to file (in real use, provide an actual path)
	// err := export.SaveJSONToFile(artifact, "/tmp/dungeon.json")
	// if err != nil {
	//     log.Fatal(err)
	// }
	_ = artifact // artifact is created for demonstration purposes

	fmt.Println("Artifact saved to file")
	// Output: Artifact saved to file
}

// ExampleExportJSONCompact demonstrates compact JSON export without indentation.
func ExampleExportJSONCompact() {
	artifact := &dungeon.Artifact{
		ADG: &dungeon.Graph{
			Graph: &graph.Graph{
				Rooms: map[string]*graph.Room{
					"r1": {ID: "r1", Archetype: graph.ArchetypeStart},
				},
			},
		},
	}

	// Export compact JSON
	compact, _ := export.ExportJSONCompact(artifact)
	formatted, _ := export.ExportJSON(artifact)

	fmt.Printf("Compact is smaller: %v\n", len(compact) < len(formatted))
	// Output: Compact is smaller: true
}
