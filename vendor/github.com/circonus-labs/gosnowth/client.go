package gosnowth

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// SnowthNode - The representation of a snowth node. An IRONdb cluster consists of
// several nodes.  A SnowthNode has a URL to the API of that Node, an identifier,
// and a current topology.  The identifier is how the node is identified within
// the cluster, and the topology is the current topology that the node falls
// within.  A topology is a set of nodes that distribute data amoungst each other.
type SnowthNode struct {
	url             *url.URL
	identifier      string
	currentTopology string
}

// GetURL - This will return the *url.URL of the given SnowthNode.  This will be
// useful if you need the raw connection string of a given snowthnode, such as in
// the event you are making a proxy for a snowth node.
func (sn *SnowthNode) GetURL() *url.URL {
	return sn.url
}

// GetCurrentTopology - This will return the hash string representation of the
// node's current topology.
func (sn *SnowthNode) GetCurrentTopology() string {
	return sn.currentTopology
}

// httpClient - interface in order to mock http requests
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// SnowthClient - The client functionality for operating against SnowthNodes.
// Operations for the client can be broken down into 6 major sections:
//		1.) State and Topology
// Within the state and topology APIs, there are several useful apis, including
// apis to retrieve Node state, Node gossip information, topology information,
// and toporing information.  Each of these operations is implemented as a method
// on this client.
//		2.) Rebalancing APIs
// In order to add or remove nodes within an IRONdb cluster you will have to use
// the rebalancing APIs.  Implemented within this package you will be able to
// load a new topology, rebalance nodes to the new topology, as well as check
// load state information and abort a topology change.
//		3.) Data Retrieval APIs
// IRONdb houses data, and the data retrieval APIs allow for accessing of that
// stored data.  Data types implemented include NNT, Text, and Histogram data
// element types.
//		4.) Data Submission APIs
// IRONdb houses data, to which you can use to submit data to the cluster.  Data
// types supported include the same for the retrieval APIs, NNT, Text and
// Histogram data types.
//		5.) Data Deletion APIs
// Data sometimes needs to be deleted, and that is performed with the data
// deletion APIs.  This client implements the data deletion apis to remove data
// from the nodes.
//		6.) Lua Extensions APIs
type SnowthClient struct {
	c httpClient

	// in order to keep track of healthy nodes within the cluster,
	// we have two lists of SnowthNode types, active and inactive.
	activeNodesMu *sync.RWMutex
	activeNodes   []*SnowthNode

	inactiveNodesMu *sync.RWMutex
	inactiveNodes   []*SnowthNode

	// watchInterval is the duration between checks to tell if a node is active
	// or inactive.
	watchInterval time.Duration
}

// NewSnowthClient - given a variadic addrs parameter, the client will
// construct all the needed state to communicate with a group of nodes
// which constitute a cluster.  It will return a pointer to a SnowthClient.
// The discover parameter when true will allow the client to discover new
// nodes from the topology
func NewSnowthClient(discover bool, addrs ...string) (*SnowthClient, error) {
	sc := &SnowthClient{
		c:               http.DefaultClient,
		activeNodesMu:   new(sync.RWMutex),
		activeNodes:     []*SnowthNode{},
		inactiveNodesMu: new(sync.RWMutex),
		inactiveNodes:   []*SnowthNode{},
		watchInterval:   5 * time.Second,
	}

	// for each of the addrs we need to parse the connection string,
	// then create a node for that connection string, poll the state
	// of that node, and populate the identifier and topology of that
	// node.  Finally we will add the node and activate it.
	for _, addr := range addrs {
		url, err := url.Parse(addr)
		if err != nil {
			// this node had an error, put on inactive list
			log.Printf("failed to bootstrap state of node: %+v", err)
			continue
		}
		node := &SnowthNode{url: url}
		// call get state to populate the id of this node
		state, err := sc.GetNodeState(node)
		if err != nil {
			// this node had an error, put on inactive list
			log.Printf("failed to bootstrap state of node: %+v", err)
			continue
		}
		node.identifier = state.Identity
		node.currentTopology = state.Current
		sc.AddNodes(node)
		sc.ActivateNodes(node)
	}
	// start a goroutine to watch for changes in state of the nodes,
	// and manage the active/inactive lists accordingly
	go sc.watchAndUpdate()

	if discover {
		// for robustness, we will perform a discovery of associated nodes
		// this works by pulling the topology information for given nodes
		// and adding nodes discovered within the topology into the client
		if err := sc.discoverNodes(); err != nil {
			log.Printf("failed to perform discovery of new nodes")
		}
	}

	return sc, nil
}

// isNodeActive - The check to see if a given node is active or not.
// this will take into account ability to get the node state, gossip
// information as well as the gossip age of the node.  If the age is
// larger than 10 we will not consider this node active.
func (sc *SnowthClient) isNodeActive(node *SnowthNode) bool {
	var id = node.identifier
	if id == "" {
		// go get state to figure out identity
		state, err := sc.GetNodeState(node)
		if err != nil {
			// error means we failed, node is not active
			return false
		}
		id = state.Identity
	}
	gossip, err := sc.GetGossipInfo(node)
	if err != nil {
		return false
	}
	var age float64 = 100.0
	for _, entry := range []GossipDetail(*gossip) {
		if entry.ID == id {
			age = entry.Age
			break
		}
	}
	if age > 10.0 {
		return false
	}
	return true
}

// watchAndUpdate - watch gossip data for all nodes, and move the nodes to active
// or inactive as required.  Will walk through the inactive nodes, checking for
// aliveness, then walk through active nodes checking for aliveness.
func (sc *SnowthClient) watchAndUpdate() {
	for {
		<-time.After(sc.watchInterval)
		for _, node := range sc.ListInactiveNodes() {
			if sc.isNodeActive(node) {
				// move to active
				sc.ActivateNodes(node)
			}
		}
		for _, node := range sc.ListActiveNodes() {
			if !sc.isNodeActive(node) {
				// move to active
				sc.DeactivateNodes(node)
			}
		}
	}
}

// discoverNodes - private method for the client to discover peer nodes
// related to the topology.  This function will go through the active nodes
// get the topology information which shows all other nodes included in
// the topology, and adds them as snowth nodes to this client's active pool
// of nodesh
func (sc *SnowthClient) discoverNodes() error {
	// take our list of active nodes, interrogate gossipinfo
	// get more nodes from the gossip info
	var (
		success = false
		mErr    = newMultiError()
	)
	for _, node := range sc.ListActiveNodes() {
		// lookup the topology
		topology, err := sc.GetTopologyInfo(node)
		if err != nil {
			mErr.Add(errors.Wrap(err, "error getting topology info: %+v"))
			continue
		}

		// populate all the nodes with the appropriate topology information
		for _, topoNode := range topology.Nodes {
			sc.populateNodeInfo(node.GetCurrentTopology(), topoNode)
		}
		success = true
	}

	if !success {
		// we didn't get any topology information, therefore we didn't
		// discover correctly, return the multitude of errors
		return mErr
	}
	return nil
}

// populateNodeInfo - this helper method populates an existing node with the
// details from the topology.  If a node doesn't exist, it will be added
// to the list of active nodes in the client.
func (sc *SnowthClient) populateNodeInfo(hash string, topology TopologyNode) {
	var found = false

	sc.activeNodesMu.Lock()
	for i := 0; i < len(sc.activeNodes); i++ {
		if sc.activeNodes[i].identifier == topology.ID {
			found = true
			url := url.URL{
				Scheme: "http",
				Host:   fmt.Sprintf("%s:%d", topology.Address, topology.APIPort),
			}
			sc.activeNodes[i].url = &url
			sc.activeNodes[i].currentTopology = hash
			continue
		}
	}
	sc.activeNodesMu.Unlock()
	sc.inactiveNodesMu.Lock()
	for i := 0; i < len(sc.inactiveNodes); i++ {
		found = true
		if sc.inactiveNodes[i].identifier == topology.ID {
			url := url.URL{
				Scheme: "http",
				Host:   fmt.Sprintf("%s:%d", topology.Address, topology.APIPort),
			}
			sc.inactiveNodes[i].url = &url
			sc.inactiveNodes[i].currentTopology = hash
			continue
		}
	}
	sc.inactiveNodesMu.Unlock()
	if !found {
		newNode := &SnowthNode{
			identifier: topology.ID,
			url: &url.URL{
				Scheme: "http",
				Host:   fmt.Sprintf("%s:%d", topology.Address, topology.APIPort),
			},
			currentTopology: hash,
		}
		sc.AddNodes(newNode)
		sc.ActivateNodes(newNode)
	}
}

// doChangeActivation - perform an activation state change
func (sc *SnowthClient) doChangeActivation(from, to *[]*SnowthNode, nodes []*SnowthNode) {
	sc.activeNodesMu.Lock()
	defer sc.activeNodesMu.Unlock()
	sc.inactiveNodesMu.Lock()
	defer sc.inactiveNodesMu.Unlock()
	for _, v := range nodes {
		moveNode(from, to, v)
	}
}

// ActivateNodes - given a list of nodes, make said nodes active for the client
func (sc *SnowthClient) ActivateNodes(nodes ...*SnowthNode) {
	sc.doChangeActivation(&sc.inactiveNodes, &sc.activeNodes, nodes)
}

// DeactivateNodes - given a list of nodes, make said nodes inactive
func (sc *SnowthClient) DeactivateNodes(nodes ...*SnowthNode) {
	sc.doChangeActivation(&sc.activeNodes, &sc.inactiveNodes, nodes)
}

// AddNodes - add nodes parameters to the inactive node list
func (sc *SnowthClient) AddNodes(nodes ...*SnowthNode) {
	sc.inactiveNodesMu.Lock()
	defer sc.inactiveNodesMu.Unlock()
	sc.inactiveNodes = append(sc.inactiveNodes, nodes...)
}

// doListNodes - helper to list the nodes, active or inactive
func doListNodes(nodes *[]*SnowthNode, mu *sync.RWMutex) []*SnowthNode {
	mu.RLock()
	defer mu.RUnlock()
	var result = []*SnowthNode{}
	for _, url := range *nodes {
		result = append(result, url)
	}
	return result
}

// ListInactiveNodes - list all of the currently inactive nodes
func (sc *SnowthClient) ListInactiveNodes() []*SnowthNode {
	return doListNodes(&sc.inactiveNodes, sc.inactiveNodesMu)
}

// ListActiveNodes - list all of the currently active nodes
func (sc *SnowthClient) ListActiveNodes() []*SnowthNode {
	return doListNodes(&sc.activeNodes, sc.activeNodesMu)
}

// do - helper to perform the request for the client
func (sc *SnowthClient) do(node *SnowthNode, method, url string,
	body io.Reader, respValue interface{},
	decodeFunc func(interface{}, io.Reader) error) error {

	r, err := http.NewRequest(method, sc.getURL(node, url), body)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	resp, err := sc.c.Do(r)
	if err != nil {
		return errors.Wrap(err, "failed to perform request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("non-success status code returned: %s -> %s",
			resp.Status, string(body))
	}

	if respValue != nil {
		if err := decodeFunc(respValue, resp.Body); err != nil {
			return errors.Wrap(err, "failed to decode")
		}
	}

	return nil
}

// getURL - helper to resolve a reference against a particular node
func (sc *SnowthClient) getURL(node *SnowthNode, ref string) string {
	return resolveURL(node.url, ref)
}
