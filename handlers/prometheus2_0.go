package handlers

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/prompb"
	uuid "github.com/satori/go.uuid"
)

func handleRequest(ctx echo.Context, req proto.Message, accountID *int32, checkName *string, checkUUID *uuid.UUID) error {
	// close request body
	defer ctx.Request().Body.Close()
	var (
		// create our prometheus format decoder
		err error
	)
	// validation of url parameters
	*accountID, err = ValidateAccountID(ctx)
	if err != nil {
		// 400, invalid account id
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account id in URL")
	}
	*checkUUID, err = ValidateCheckUUID(ctx)
	if err != nil {
		// 400, invalid account id
		return echo.NewHTTPError(http.StatusBadRequest, "invalid check_uuid in URL")
	}
	*checkName, err = ValidateCheckName(ctx)
	if err != nil {
		// 400, invalid account id
		return echo.NewHTTPError(http.StatusBadRequest, "invalid check_name in URL")
	}

	// pull the body off of the request into a byte slice
	compressed, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.Logger().Errorf("failed to read request body: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "could not read request body")
	}

	reqBuf, err := snappy.Decode(nil, compressed)
	if err != nil {
		ctx.Logger().Errorf("failed to decompress request body: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode request body")
	}

	if err := proto.Unmarshal(reqBuf, req); err != nil {
		ctx.Logger().Errorf("failed to decode protobuf request: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode request body")
	}
	return nil
}

// PrometheusWrite2_0 - the application handler which converts a prometheus
// MetricFamily message to a MetricList for ingestion into IRONdb
func PrometheusWrite2_0(ctx echo.Context) error {
	// decode, read and deserialize the prometheus request
	var (
		req          prompb.WriteRequest
		accountID    int32
		checkName    string
		checkUUID    uuid.UUID
		snowthClient SnowthClientI
	)
	if client, ok := ctx.Get("snowthClient").(SnowthClientI); ok {
		snowthClient = client
	}
	if err := handleRequest(ctx, &req, &accountID, &checkName, &checkUUID); err != nil {
		return err
	}

	// make the metric list flatbuffer data
	metricList, err := MakeMetricList(req.GetTimeseries(), accountID, checkName, checkUUID)
	if err != nil {
		ctx.Logger().Errorf("failed to convert to flatbuffer: %s", err.Error())
		return err
	}

	// pull a random snowth node from the client to send request to
	node := getRandomNode(snowthClient.ListActiveNodes()...)
	if node == nil {
		// we are degraded, there are no active nodes
		ctx.Logger().Errorf("there are no active nodes... active: %+v, inactive: %+v\n", snowthClient.ListActiveNodes(), snowthClient.ListInactiveNodes())
		return errors.New("no active irondb nodes")
	}
	ctx.Logger().Debugf("using node: %s of %+v", node.GetURL(), snowthClient.ListActiveNodes())
	fmt.Println(hex.Dump(metricList))
	// perform the write to IRONdb
	if err = snowthClient.WriteRaw(node, bytes.NewBuffer(metricList), true, uint64(len(req.GetTimeseries()))); err != nil {
		ctx.Logger().Errorf("failed to write flatbuffer: %s", err.Error())
		return errors.Wrap(err, "failed to write flatbuffer")
	}
	ctx.Logger().Debugf("converted flatbuffer: %+v\n", metricList)
	return nil
}

// PrometheusRead2_0 - the application handler which converts a prometheus
// read message to an IRONdb read message, and returns the results converted
// back into prometheus output
func PrometheusRead2_0(ctx echo.Context) error {
	// decode, read and deserialize the prometheus request
	var (
		req          prompb.ReadRequest
		resp         prompb.ReadResponse
		accountID    int32
		checkName    string
		checkUUID    uuid.UUID
		snowthClient SnowthClientI
	)
	if client, ok := ctx.Get("snowthClient").(SnowthClientI); ok {
		snowthClient = client
	}
	if err := handleRequest(ctx, &req, &accountID, &checkName, &checkUUID); err != nil {
		return err
	}
	// pull a random snowth node from the client to send request to
	node := getRandomNode(snowthClient.ListActiveNodes()...)
	if node == nil {
		// we are degraded, there are no active nodes
		ctx.Logger().Errorf("there are no active nodes... active: %+v, inactive: %+v\n", snowthClient.ListActiveNodes(), snowthClient.ListInactiveNodes())
		return errors.New("no active irondb nodes")
	}
	ctx.Logger().Debugf("using node: %s of %+v", node.GetURL(), snowthClient.ListActiveNodes())

	// TODO: perform the read...

	data, err := proto.Marshal(&resp)
	if err != nil {
		ctx.Logger().Errorf("failed to marshal response: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to marshal response")
	}

	ctx.Response().Header().Set("Content-Type", "application/x-protobuf")
	ctx.Response().Header().Set("Content-Encoding", "snappy")
	ctx.Response().WriteHeader(http.StatusOK)

	var compressed = snappy.Encode(nil, data)
	if _, err := ctx.Response().Write(compressed); err != nil {
		ctx.Logger().Errorf("failed to write response: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to write response")
	}

	return nil
}
