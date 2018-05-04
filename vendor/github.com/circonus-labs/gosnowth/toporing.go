package gosnowth

import (
	"encoding/xml"
	"path"
)

// GetTopoRingInfo - Get the toporing information from the node.
func (sc *SnowthClient) GetTopoRingInfo(hash string, node *SnowthNode) (toporing *TopoRing, err error) {
	toporing = new(TopoRing)
	err = sc.do(node, "GET", path.Join("/toporing/xml", hash), nil, toporing, decodeXMLFromResponse)
	return
}

// TopoRing - structure for the response of the toporing api calls
type TopoRing struct {
	XMLName      xml.Name         `xml:"vnodes" json:"-"`
	VirtualNodes []TopoRingDetail `xml:"vnode"`
	NumberNodes  int              `xml:"n,attr" json:"-"`
}

// TopoRingDetail - detail node information from toporing api call
type TopoRingDetail struct {
	XMLName  xml.Name `xml:"vnode" json:"-"`
	ID       string   `xml:"id,attr" json:"id"`
	IDX      int      `xml:"idx,attr" json:"idx"`
	Location float64  `xml:"location,attr" json:"location"`
}
