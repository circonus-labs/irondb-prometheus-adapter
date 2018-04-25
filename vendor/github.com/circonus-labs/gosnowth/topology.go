package gosnowth

import (
	"encoding/xml"
	"path"

	"github.com/pkg/errors"
)

// GetTopologyInfo - Get the topology information from the node.
func (sc *SnowthClient) GetTopologyInfo(node *SnowthNode) (topology *Topology, err error) {
	topology = new(Topology)
	err = sc.do(node, "GET",
		path.Join("/topology/xml", node.GetCurrentTopology()),
		nil, topology, decodeXMLFromResponse)
	return
}

// LoadTopology - Load a new topology. Will not activate, just load and store.
func (sc *SnowthClient) LoadTopology(hash string, topology *Topology, node *SnowthNode) (err error) {
	reqBody, err := encodeXML(topology)
	if err != nil {
		return errors.Wrap(err, "failed to encode request data")
	}
	err = sc.do(node, "POST", path.Join("/topology", hash), reqBody, nil, nil)
	return
}

// ActivateTopology - Switch to a new topology.  THIS IS DANGEROUS.
func (sc *SnowthClient) ActivateTopology(hash string, node *SnowthNode) (err error) {
	err = sc.do(node, "GET", path.Join("/activate", hash), nil, nil, nil)
	return
}

// Topology - the topology structure from the API
type Topology struct {
	XMLName     xml.Name       `xml:"nodes" json:"-"`
	NumberNodes int            `xml:"n,attr" json:"-"`
	Hash        string         `xml:"-"`
	Nodes       []TopologyNode `xml:"node"`
}

// TopologyNode - the topology node structure from the API
type TopologyNode struct {
	XMLName     xml.Name `xml:"node" json:"-"`
	ID          string   `xml:"id,attr" json:"id"`
	Address     string   `xml:"address,attr" json:"address"`
	Port        uint16   `xml:"port,attr" json:"port"`
	APIPort     uint16   `xml:"apiport,attr" json:"apiport"`
	Weight      int      `xml:"weight,attr" json:"weight"`
	NumberNodes int      `xml:"-" json:"n"`
}
