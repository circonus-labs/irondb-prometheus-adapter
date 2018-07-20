package gosnowth

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

const (
	FlatbufferContentType = "application/x-circonus-metric-list-flatbuffer"
)

// WriteRaw - Write Raw data to a node, data should be a io.Reader
// and node is the node to write the data to
func (sc *SnowthClient) WriteRaw(node *SnowthNode, data io.Reader, fb bool, dataPoints uint64) (err error) {
	id := uuid.NewV4()
	r, err := http.NewRequest("POST", sc.getURL(node, "/raw"), data)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	r.Header.Add("X-Snowth-Datapoints", strconv.FormatUint(dataPoints, 10))
	// is flatbuffer?
	if fb {
		r.Header.Add("Content-Type", FlatbufferContentType)
	}

	var start = time.Now()
	resp, err := sc.c.Do(r)
	if err != nil {
		return errors.Wrap(err, "failed to perform request")
	}
	defer resp.Body.Close()

	sc.Logger.Debugf("Snowth Response: %s %+v", id.String(), resp)
	sc.Logger.Debugf("Snowth Response Latency: %s %+v", id.String(), time.Now().Sub(start))

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("non-success status code returned: %s -> %s",
			resp.Status, string(body))
	}
	return
}
