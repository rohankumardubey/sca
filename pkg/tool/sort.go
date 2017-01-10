package tool

import (
	"encoding/json"
	"sort"

	"github.com/docker/docker/api/types/swarm"
	docker "github.com/fsouza/go-dockerclient"
)

//Min tools minimum of int
func Min(A, B int) int {
	min := A
	if A > B {
		min = B
	}
	return min
}

//SortedKeys tools sort map[string]
func SortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

//ByCID sort class
type ByCID []docker.APIContainers

func (b ByCID) Len() int           { return len(b) }
func (b ByCID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByCID) Less(i, j int) bool { return b[i].ID < b[j].ID }

//ByNID sort class
type ByNID []docker.Network

func (b ByNID) Len() int           { return len(b) }
func (b ByNID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByNID) Less(i, j int) bool { return b[i].ID < b[j].ID }

//ByIID sort class
type ByIID []docker.APIImages

func (b ByIID) Len() int           { return len(b) }
func (b ByIID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByIID) Less(i, j int) bool { return b[i].ID < b[j].ID }

//ByVName sort class
type ByVName []docker.Volume

func (b ByVName) Len() int           { return len(b) }
func (b ByVName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByVName) Less(i, j int) bool { return b[i].Name < b[j].Name }

//ByPort sort class
type ByPort []docker.APIPort

func (b ByPort) Len() int      { return len(b) }
func (b ByPort) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByPort) Less(i, j int) bool {
	b1, _ := json.Marshal(b[i])
	b2, _ := json.Marshal(b[j])
	return string(b1) < string(b2)
}

//ByMount sort class
type ByMount []docker.APIMount

func (b ByMount) Len() int      { return len(b) }
func (b ByMount) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByMount) Less(i, j int) bool {
	b1, _ := json.Marshal(b[i])
	b2, _ := json.Marshal(b[j])
	return string(b1) < string(b2)
}

//ByPeer sort class
type ByPeer []swarm.Peer

func (b ByPeer) Len() int      { return len(b) }
func (b ByPeer) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByPeer) Less(i, j int) bool {
	b1, _ := json.Marshal(b[i])
	b2, _ := json.Marshal(b[j])
	return string(b1) < string(b2)
}
