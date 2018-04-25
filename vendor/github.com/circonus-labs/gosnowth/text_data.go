package gosnowth

import (
	"bytes"
	"encoding/json"
	"path"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// WriteText - Write Text data to a node, data should be a slice of TextData
// and node is the node to write the data to
func (sc *SnowthClient) WriteText(node *SnowthNode, data ...TextData) (err error) {
	var (
		buf = new(bytes.Buffer)
		enc = json.NewEncoder(buf)
	)
	if err := enc.Encode(data); err != nil {
		return errors.Wrap(err, "failed to encode TextData for write")
	}
	err = sc.do(node, "POST", "/write/text", buf, nil, nil)
	return
}

func (sc *SnowthClient) ReadTextValues(
	node *SnowthNode, start, end time.Time,
	id, metric string) ([]TextValue, error) {
	var (
		tvr = new(TextValueResponse)
		err = sc.do(node, "GET", path.Join("/read",
			strconv.FormatInt(start.Unix(), 10),
			strconv.FormatInt(end.Unix(), 10),
			id, metric), nil, tvr, decodeJSONFromResponse)
	)

	return tvr.Data, err
}

type TextValueResponse struct {
	Data []TextValue
}

func (tvr *TextValueResponse) UnmarshalJSON(b []byte) error {
	tvr.Data = []TextValue{}
	var values = [][]interface{}{}

	if err := json.Unmarshal(b, &values); err != nil {
		return errors.Wrap(err, "failed to deserialize nnt average response")
	}

	for _, entry := range values {
		var tv = TextValue{}
		tv.Value = entry[1].(string)
		// grab the timestamp
		if v, ok := entry[0].(float64); ok {
			tv.Time = time.Unix(int64(v), 0)
		}
		tvr.Data = append(tvr.Data, tv)
	}
	return nil
}

type TextValue struct {
	Time  time.Time
	Value string
}

// TextData - representation of Text Data for data submission and retrieval
type TextData struct {
	Metric string `json:"metric"`
	ID     string `json:"id"`
	Offset string `json:"offset"`
	Value  string `json:"value"`
}
