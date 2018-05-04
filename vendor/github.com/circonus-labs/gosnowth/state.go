package gosnowth

import (
	"encoding/json"
	"strings"
)

// GetNodeState - Get the node state from the client.
func (sc *SnowthClient) GetNodeState(node *SnowthNode) (state *NodeState, err error) {
	state = new(NodeState)
	err = sc.do(node, "GET", "/state", nil, state, decodeJSONFromResponse)
	return
}

// NodeState - the structure of the /state api call
type NodeState struct {
	Identity      string   `json:"identity"`
	Current       string   `json:"current"`
	Next          string   `json:"next"`
	NNT           Rollup   `json:"nnt"`
	Text          Rollup   `json:"text"`
	Histogram     Rollup   `json:"histogram"`
	BaseRollup    uint64   `json:"base_rollup"`
	Rollups       []uint64 `json:"rollups"`
	NNTCacheSize  uint64   `json:"nnt_cache_size"`
	RUsageUTime   float64  `json:"rusage.utime"`
	RUsageSTime   float64  `json:"rusage.stime"`
	RUsageMaxRSS  uint64   `json:"rusage.maxrss"`
	RUsageMinFLT  uint64   `json:"rusage.minflt"`
	RUsageMajFLT  uint64   `json:"rusage.majflt"`
	RUsageNSwap   uint64   `json:"rusage.nswap"`
	RUsageInBlock uint64   `json:"rusage.inblock"`
	RUsageOuBlock uint64   `json:"rusage.oublock"`

	RUsageMsgSnd   uint64  `json:"rusage.msgsnd"`
	RUsageMsgRcv   uint64  `json:"rusage.msgrcv"`
	RUsageNSignals uint64  `json:"rusage.nsignals"`
	RUsageNvcSW    uint64  `json:"rusage.nvcsw"`
	RUsageNivcSW   uint64  `json:"rusage.nivcsw"`
	MaxPeerLag     float64 `json:"max_peer_lag"`
	AvgPeerLag     float64 `json:"avg_peer_lag"`

	Features Features `json:"features"`

	Version     string `json:"version"`
	Application string `json:"application"`
}

// Rollup - the structure that defines the rollup (nnt,text,histogram) from api
type Rollup struct {
	RollupEntries
	RollupList []uint32      `json:"rollups"`
	Aggregate  RollupDetails `json:"aggregate"`
}

// UnmarshalJSON - we need a custom unmarshal so that we can have the
// dynamic rollup key/value pairs
func (r *Rollup) UnmarshalJSON(b []byte) error {
	m := make(map[string]interface{})
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	if rollups, ok := m["rollups"].([]interface{}); ok {
		for _, v := range rollups {
			r.RollupList = append(r.RollupList, uint32(v.(float64)))
			delete(m, "rollup")
		}
	}
	if aggregate, ok := m["aggregate"].(RollupDetails); ok {
		r.Aggregate = aggregate
		delete(m, "aggregate")
	}

	rr := make(map[string]RollupDetails)
	for k, v := range m {
		if strings.HasPrefix(k, "rollup_") {
			b, _ := json.Marshal(v)
			rd := new(RollupDetails)
			json.Unmarshal(b, rd)
			rr[k] = *rd
		}
	}

	r.RollupEntries = RollupEntries(rr)
	return nil
}

// RollupEntries - this is the dynamic map of string to rollups in the state api
type RollupEntries map[string]RollupDetails

// RollupDetails - the details included in the rollup
type RollupDetails struct {
	FilesSystem   FileSystemDetails `json:"fs"`
	PutCalls      uint64            `json:"put.calls"`
	PutElapsedUS  uint64            `json:"put.elapsed_us"`
	GetCalls      uint64            `json:"get.calls"`
	GetProxyCalls uint64            `json:"get.proxy_calls"`
	GetCount      uint64            `json:"get.count"`
	GetElapsedUS  uint64            `json:"get.elapsed_us"`
	ExtendCalls   uint64            `json:"extend.calls"`
}

// FileSystemDetails - details about the filesystem from the state api call
type FileSystemDetails struct {
	ID      uint64  `json:"id"`
	TotalMB float64 `json:"totalMb"`
	FreeMB  float64 `json:"availMb"`
}

// Features - these are the features supported by the node
type Features struct {
	TextStore               bool `json:"text:store"`
	HistogramStore          bool `json:"histogram:store"`
	NNTSecondOrder          bool `json:"nnt:second_order"`
	HistogramDynamicRollups bool `json:"hisogram:dynamic_rollups"`
	NNTStore                bool `json:"nnt:store"`
	FeatureFlags            bool `json:"features"`
}

// UnmarshalJSON - conversion from the string 1/0 representation to bool
func (f *Features) UnmarshalJSON(b []byte) error {
	f.TextStore = false
	f.HistogramStore = false
	f.NNTSecondOrder = false
	f.HistogramDynamicRollups = false
	f.NNTStore = false
	f.FeatureFlags = false

	m := make(map[string]string)
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	for k, v := range m {
		switch k {
		case "text:store":
			if v == "1" {
				f.TextStore = true
			}
			break
		case "histogram:store":
			if v == "1" {
				f.HistogramStore = true
			}
			break
		case "nnt:second_order":
			if v == "1" {
				f.NNTSecondOrder = true
			}
			break
		case "histogram:dynamic_rollups":
			if v == "1" {
				f.HistogramDynamicRollups = true
			}
			break
		case "nnt:store":
			if v == "1" {
				f.NNTStore = true
			}
			break
		case "features":
			if v == "1" {
				f.FeatureFlags = true
			}
			break
		}
	}
	return nil
}
