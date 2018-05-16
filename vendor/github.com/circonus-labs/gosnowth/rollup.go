package gosnowth

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"
)

type RollupValues struct {
	Timestamp int64
	Value     float64
}

func (rv *RollupValues) UnmarshalJSON(b []byte) error {
	tt := []interface{}{&rv.Timestamp, &rv.Value}
	json.Unmarshal(b, &tt)
	if len(tt) < 2 { // error not enough fields
		return fmt.Errorf("rollup value should contain two entries, %d given in payload", len(tt))
	}
	return nil
}

// ReadRollupValues - Read Rollup data from a node
func (sc *SnowthClient) ReadRollupValues(
	node *SnowthNode, id, metric string, tags []string, rollup time.Duration, start, end time.Time) ([]RollupValues, error) {

	var (
		start_ts = start.Unix() - start.Unix()%int64(rollup/time.Second)
		end_ts   = end.Unix() - end.Unix()%int64(rollup/time.Second) + int64(rollup/time.Second)
	)

	urlencodedTags := url.QueryEscape(strings.Join(tags, ","))

	var (
		r   = []RollupValues{}
		err = sc.do(node, "GET", fmt.Sprintf(
			"%s?stream_tags=%s&start_ts=%d&end_ts=%d&rollup_span=%ds",
			path.Join("/rollup", id, metric), urlencodedTags, start_ts, end_ts,
			int(rollup/time.Second)), nil, &r, decodeJSONFromResponse)
	)
	return r, err
}
