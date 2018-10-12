package handlers

import (
	"fmt"
	"testing"
)

/*
func TestPrometheusWrite2_0(t *testing.T) {
	// setup echo bits
	e := echo.New()

	promMsg := prompb.WriteRequest{
		Timeseries: []*prompb.TimeSeries{
			&prompb.TimeSeries{
				Samples: []*prompb.Sample{
					&prompb.Sample{
						Timestamp: int64(time.Now().Unix()),
						Value:     42,
					},
				},
				Labels: []*prompb.Label{
					&prompb.Label{
						Name: "label-name", Value: "label-value",
					},
				},
			},
		},
	}

	data, err := proto.Marshal(&promMsg)
	if err != nil {
		t.Error("failed to marshal prompb message: ", err.Error())
	}
	var postBody []byte
	postBody = snappy.Encode(nil, data)
	if err != nil {
		t.Error("failed to compress prompb message: ", err.Error())
	}

	// mock snowth client
	snowthClient := new(mockSnowthClient)
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// add the snowth client to the base context
			ctx.Set("snowthClient", snowthClient)
			return next(ctx)
		}
	})
	e.POST("/prometheus/2.0/write/:account/:check_uuid/:check_name", PrometheusWrite2_0)

	url := fmt.Sprintf("/prometheus/2.0/write/42/%s/check_name", uuid.NewV4().String())
	r, _ := http.NewRequest("POST", url, bytes.NewBuffer(postBody))
	w := httptest.NewRecorder()

	e.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("failure in write request: %s %d\n", w.Body.String(), w.Code)
	}
}

func TestPrometheusRead2_0(t *testing.T) {
	// setup echo bits
	e := echo.New()

	promMsg := prompb.ReadRequest{
		Queries: []*prompb.Query{
			&prompb.Query{
				Matchers: []*prompb.LabelMatcher{
					&prompb.LabelMatcher{
						Type:  prompb.LabelMatcher_EQ,
						Name:  "metric_name",
						Value: "metric_value",
					},
					&prompb.LabelMatcher{
						Type:  prompb.LabelMatcher_NEQ,
						Name:  "metric_name",
						Value: "metric_value",
					},
					&prompb.LabelMatcher{
						Type:  prompb.LabelMatcher_RE,
						Name:  "metric_name",
						Value: "metric_value",
					},
					&prompb.LabelMatcher{
						Type:  prompb.LabelMatcher_NRE,
						Name:  "metric_name",
						Value: "metric_value",
					},
				},
			},
		},
	}

	data, err := proto.Marshal(&promMsg)
	if err != nil {
		t.Error("failed to marshal prompb message: ", err.Error())
	}
	var postBody = snappy.Encode(nil, data)
	if err != nil {
		t.Error("failed to compress prompb message: ", err.Error())
	}

	snowthClient := new(mockSnowthClient)
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// add the snowth client to the base context
			ctx.Set("snowthClient", snowthClient)
			return next(ctx)
		}
	})
	e.POST("/prometheus/2.0/read/:account/:check_uuid/:check_name", PrometheusRead2_0)

	url := fmt.Sprintf("/prometheus/2.0/read/42/%s/check_name", uuid.NewV4().String())

	r, _ := http.NewRequest("POST", url, bytes.NewBuffer(postBody))
	w := httptest.NewRecorder()

	e.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("failure in write request: %s %d\n", w.Body.String(), w.Code)
	}

}
*/
func TestMetricNameToLabelPairs(t *testing.T) {

	metricName := `up|ST[cmdb_shard:test,cmdb_status:ALLOCATED,data_center:dal06,environment:prod,instance:prometheus-test-000-g5.prod.dal06.:9201,job:prometheus-remote-storage,monitor:test,replica:prometheus-test-000-g5.prod.dal06,tier:prometheus]`

	fmt.Printf("%+v\n", metricNameToLabelPairs(metricName))

}
