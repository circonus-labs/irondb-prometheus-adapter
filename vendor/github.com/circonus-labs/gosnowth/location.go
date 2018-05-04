package gosnowth

import (
	"path"
)

// LocateMetric - locate which nodes a metric lives on
func (sc *SnowthClient) LocateMetric(uuid string, metric string, node *SnowthNode) (location *DataLocation, err error) {
	location = new(DataLocation)
	err = sc.do(node, "GET", path.Join("/locate/xml", uuid, metric), nil, location, decodeXMLFromResponse)
	return
}

// DataLocation is from the location api and mimics the topology response
type DataLocation Topology
