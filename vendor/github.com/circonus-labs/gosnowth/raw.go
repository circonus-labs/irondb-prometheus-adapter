package gosnowth

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	FlatbufferContentType = "application/x-circonus-metric-list-flatbuffer"
)

// WriteRaw - Write Raw data to a node, data should be a io.Reader
// and node is the node to write the data to
func (sc *SnowthClient) WriteRaw(node *SnowthNode, data io.Reader, fb bool, dataPoints uint64) (err error) {

	r, err := http.NewRequest("POST", sc.getURL(node, "/raw"), data)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	r.Header.Add("X-Snowth-Datapoints", strconv.FormatUint(dataPoints, 10))
	// is flatbuffer?
	if fb {
		r.Header.Add("Content-Type", FlatbufferContentType)
	}

	sc.Logger.Debugf("Snowth Request: %+v", r)

	var start = time.Now()
	resp, err := sc.c.Do(r)
	if err != nil {
		return errors.Wrap(err, "failed to perform request")
	}
	defer resp.Body.Close()

	sc.Logger.Debugf("Snowth Response: %+v", resp)
	sc.Logger.Debugf("Snowth Response Latency: %+v", time.Now().Sub(start))

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		sc.Logger.Warnf("status code not 200: %+v", resp)
		return fmt.Errorf("non-success status code returned: %s -> %s",
			resp.Status, string(body))
	}
	return
}
