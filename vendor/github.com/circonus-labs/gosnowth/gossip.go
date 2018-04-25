package gosnowth

// GetGossipInfo - Get the gossip information from the client.  The gossip
// response body will include a list of "GossipDetail" which provide
// the identifier of the node, the node's gossip_time, gossip_age, as well
// as topology state, current and next topology.  This gossip information is
// useful to know because you can get availablility information about the node
func (sc *SnowthClient) GetGossipInfo(node *SnowthNode) (gossip *Gossip, err error) {
	gossip = new(Gossip)
	err = sc.do(node, "GET", "/gossip/json", nil, gossip, decodeJSONFromResponse)
	return
}

// Gossip - the gossip information from a node.  This structure includes
// information on how the nodes are communicating with each other, and if an
// nodes are behind with each other with regards to data replication.
type Gossip []GossipDetail

// GossipDetail - Gossip information about a node identified by ID
type GossipDetail struct {
	ID          string        `json:"id"`
	Time        float64       `json:"gossip_time,string"`
	Age         float64       `json:"gossip_age,string"`
	CurrentTopo string        `json:"topo_current"`
	NextTopo    string        `json:"topo_next"`
	TopoState   string        `json:"topo_state"`
	Latency     GossipLatency `json:"latency"`
}

// GossipLatency - a map of the uuid of the node to the latency in seconds
type GossipLatency map[string]string
