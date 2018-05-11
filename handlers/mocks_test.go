package handlers

import (
	"io"
	"time"

	"github.com/circonus-labs/gosnowth"
)

type mockSnowthClient struct {
	mockWriteRaw          func(*gosnowth.SnowthNode, io.Reader, bool, uint64) error
	mockFindTags          func(*gosnowth.SnowthNode, int32, string) ([]gosnowth.FindTagsItem, error)
	mockListActiveNodes   func() []*gosnowth.SnowthNode
	mockListInactiveNodes func() []*gosnowth.SnowthNode
	mockReadNNTValues     func(*gosnowth.SnowthNode, time.Time, time.Time, int64, string, string, string) ([]gosnowth.NNTValue, error)
	mockReadRollupValues  func(
		node *gosnowth.SnowthNode, id, metric string, tags []string, rollup time.Duration, start, end time.Time) ([]gosnowth.RollupValues, error)
}

func (msc *mockSnowthClient) ReadRollupValues(
	node *gosnowth.SnowthNode, id, metric string, tags []string, rollup time.Duration, start, end time.Time) ([]gosnowth.RollupValues, error) {
	if msc.mockReadRollupValues != nil {
		return msc.mockReadRollupValues(node, id, metric, tags, rollup, start, end)
	}
	return nil, nil
}

func (msc *mockSnowthClient) ReadNNTValues(
	node *gosnowth.SnowthNode, start, end time.Time, period int64,
	t, id, metric string) ([]gosnowth.NNTValue, error) {
	if msc.mockReadNNTValues != nil {
		return msc.mockReadNNTValues(node, start, end, period, t, id, metric)
	}
	return nil, nil

}

func (msc *mockSnowthClient) WriteRaw(node *gosnowth.SnowthNode, data io.Reader, fb bool, numDatapoints uint64) (err error) {
	if msc.mockWriteRaw != nil {
		return msc.mockWriteRaw(node, data, fb, numDatapoints)
	}
	return nil
}

func (msc *mockSnowthClient) FindTags(node *gosnowth.SnowthNode, acctID int32, query string) ([]gosnowth.FindTagsItem, error) {
	if msc.mockFindTags != nil {
		return msc.mockFindTags(node, acctID, query)
	}
	return nil, nil
}

func (msc *mockSnowthClient) ListActiveNodes() []*gosnowth.SnowthNode {
	if msc.mockListActiveNodes != nil {
		return msc.mockListActiveNodes()
	}
	return []*gosnowth.SnowthNode{new(gosnowth.SnowthNode)}
}
func (msc *mockSnowthClient) ListInactiveNodes() []*gosnowth.SnowthNode {
	if msc.mockListInactiveNodes != nil {
		return msc.mockListInactiveNodes()
	}
	return nil
}
