package handlers

import (
	"io"
	"math/rand"
	"time"

	"github.com/circonus-labs/gosnowth"
)

type SnowthClientI interface {
	WriteRaw(*gosnowth.SnowthNode, io.Reader, bool, uint64) error
	FindTags(*gosnowth.SnowthNode, int32, string, string, string) ([]gosnowth.FindTagsItem, error)
	ListActiveNodes() []*gosnowth.SnowthNode
	ListInactiveNodes() []*gosnowth.SnowthNode
	ReadNNTValues(*gosnowth.SnowthNode, time.Time, time.Time, int64, string, string, string) ([]gosnowth.NNTValue, error)
	ReadRollupValues(
		node *gosnowth.SnowthNode, id, metric string, tags []string, rollup time.Duration, start, end time.Time) ([]gosnowth.RollupValues, error)
}

var gen = rand.New(rand.NewSource(2))

func ChooseActiveNode(client SnowthClientI) *gosnowth.SnowthNode {
	choices := client.ListActiveNodes()
	if len(choices) == 0 {
		return nil
	}
	choice := gen.Int() % len(choices)
	return choices[choice]
}
