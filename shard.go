// shard.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson  lelawson
// Pete Wilcox      pcwilcox
// Annie Shen       ashen7
// Victoria Tran    vilatran
//
// Defines an interface and struct for the sharding system.
//

package main

import (
	"math/rand"
	"sort"
	"strings"

	"github.com/go-test/deep"
)

// Shard interface defines the interactions with the shard system
type Shard interface {
	// Returns the number of elemetns in the shard map
	CountShards() int

	// Return number of servers
	CountServers() int

	// Return true if we contain the server
	ContainsServer(string) bool

	// Return true if we contain the shard
	ContainsShard(string) bool

	// Deletes a shard ID from the shard list
	Remove(string) bool

	// Inserts an shard ID into the shard list
	Add(string) bool

	// Returns a slice of shard IDs
	GetAllShards() string

	// Returns the actual shard ID I am in
	PrimaryID() string

	// Return our IP
	GetIP() string

	// Converts the shard IDs and servers in the ID into a comma-separated string
	String() string

	// Return a random number of elements from the local view
	RandomLocal(int) []string

	// Return a random number of elements from the global view
	RandomGlobal(int) []string

	// FindBob returns a random element of a particular shard
	FindBob(string) string

	// Overwrite with a new view of the world
	Overwrite(ShardGlob)

	// GetShardGlob returns a Shard object
	GetShardGlob() ShardGlob
}

// ShardList is a struct which implements the Shard interface and holds shard ID system of servers
/*
ShardList: {
	ShardStrings: {
        A: "192.168.0.10:8081,192.168.0.10:8082",
        B: "192.168.0.10:8083,192.168.0.10:8084",
    },
    ShardSlices: {
        A: ["192.168.0.10:8081", "192.168.0.10:8082"],
        B: ["192.168.0.10:8083", "192.168.0.10:8084"],
    },
    PrimaryShard: "A",
    PrimaryIP: "192.168.0.10:8081",
}
*/
type ShardList struct {
	ShardString  map[string]string   // This is the map of shard IDs to server names
	ShardSlice   map[string][]string // this is a mapping of shard IDs to slices of server strings
	PrimaryShard string              // This is the shard ID I belong in
	PrimaryIP    string              // this is my IP
	Tree         RBTree              // This is our red-black tree holding the shard positions on the ring
	Size         int                 // total number of servers
	NumShards    int                 // total number of shards
}

// GetAllShards returns a comma-separated list of shards
func (s *ShardList) GetAllShards() string {
	if s != nil {
		var sl []string
		for k := range s.ShardString {
			sl = append(sl, k)
		}
		st := strings.Join(sl, ", ")
		return st
	}
	return ""
}

// FindBob returns a random element of the chosen shard
func (s *ShardList) FindBob(shard string) string {
	r := rand.Int()
	l := s.ShardSlice[shard]
	i := r % len(l)
	bob := l[i]
	return bob
}

// GetShardGlob returns a ShardGlob
func (s *ShardList) GetShardGlob() ShardGlob {
	if s != nil {
		g := ShardGlob{ShardList: s.ShardSlice}
		return g
	}
	return ShardGlob{}
}

// Overwrite overwrites our view of the world with another
func (s *ShardList) Overwrite(sg ShardGlob) {
	if diff := deep.Equal(sg.ShardList, s.ShardSlice); diff != nil {
		// Remove our old view of the world
		for k := range s.ShardSlice {
			delete(s.ShardSlice, k)
			delete(s.ShardString, k)
			for _, i := range getVirtualNodePositions(k) {
				s.Tree.delete(i)
			}
		}

		// Write the new one
		for k, v := range sg.ShardList {
			// Directly transfer the slices over
			s.ShardSlice[k] = v

			// Join the slices to form the string
			s.ShardString[k] = strings.Join(v, ",")

			// Check which shard we're in
			for i := range v {
				if v[i] == s.PrimaryIP {
					s.PrimaryShard = k
				}
			}

			// rebuild the tree
			for _, i := range getVirtualNodePositions(k) {
				s.Tree.put(i, k)
			}
		}

		shardChange = true
	}

}

// RandomGlobal returns a random selection of other servers from any shard
func (s *ShardList) RandomGlobal(n int) []string {
	var t []string

	if n > s.Size {
		n = s.Size - 1
	}

	for _, v := range s.ShardSlice {
		r := rand.Int() % len(v)
		if v[r] == s.PrimaryIP {
			continue
		}
		t = append(t, v[r])
		if len(t) >= n {
			break
		}
	}

	return t
}

// RandomLocal returns a random selection of other servers from within our own shard
func (s *ShardList) RandomLocal(n int) []string {
	var t []string

	l := s.ShardSlice[s.PrimaryShard]
	if n > len(l)-1 {
		n = len(l) - 2
	}

	for len(t) < n {
		r := rand.Int() % len(l)
		if l[r] == s.PrimaryIP {
			continue
		}
		t = append(t, l[r])
		if len(t) >= n {
			break
		}
	}

	return t
}

// CountServers returns the number of servers in the shard map
func (s *ShardList) CountServers() int {
	if s != nil {
		return s.Size
	}
	return 0
}

// CountShards returns the number of shards
func (s *ShardList) CountShards() int {
	if s != nil {
		return s.NumShards
	}
	return 0
}

// ContainsShard returns true if the ShardList contains a given shardID
func (s *ShardList) ContainsShard(shardID string) bool {
	if s != nil {
		_, ok := s.ShardSlice[shardID]
		return ok
	}
	return false
}

// ContainsServer checks to see if the server exists
func (s *ShardList) ContainsServer(ip string) bool {
	if s != nil {
		for _, v := range s.ShardSlice {
			for _, i := range v {
				if i == ip {
					return true
				}
			}
		}
	}
	return false
}

// Remove deletes a shard ID from the shard list
func (s *ShardList) Remove(shardID string) bool {
	if s != nil {
		delete(s.ShardString, shardID)
		delete(s.ShardSlice, shardID)
		s.NumShards--
		shardChange = true
		return true
	}
	return false
}

// Add inserts an shard ID into the my shard list
func (s *ShardList) Add(newShardID string) bool {
	if s != nil {
		// QUESTION: is here where I choose the random name, or the caller?
		// Insert newShardID into both maps
		s.ShardString[newShardID] = ""
		s.ShardSlice[newShardID] = append(s.ShardSlice[newShardID], "")
		s.NumShards++
		shardChange = true
		return true
	}
	return false
}

// PrimaryID returns the actual shard ID I am in
func (s *ShardList) PrimaryID() string {
	if s != nil {
		return s.PrimaryShard
	}
	return ""
}

// GetIP returns my IP
func (s *ShardList) GetIP() string {
	if s != nil {
		return s.PrimaryIP
	}
	return ""
}

// NumLeftoverServers returns the number of leftover servers after an uneven spread
func (s *ShardList) NumLeftoverServers() int {
	if s != nil {
		return s.Size % s.NumShards
	}
	return -1
}

// String returns a comma-separated string of shards
func (s *ShardList) String() string {
	if s != nil {
		str := make([]string, s.NumShards)

		for i := 0; i < s.NumShards; i++ {
			str[i] = shardNames[i]
		}
		j := strings.Join(str, ",")
		return j
	}
	return ""
}

// NumServerPerShard returns number of servers per shard (equally) after reshuffle
func (s *ShardList) NumServerPerShard() int {
	if s != nil {
		i := s.Size / s.NumShards
		if i >= 2 {
			return i
		}
	}
	// The caller function needs to send response to client. Insufficent shard number!!
	return -1
}

// NewShard creates a shardlist object and initializes it with the input string
func NewShard(primaryIP string, globalView string, numShards int) *ShardList {
	// init fields
	shardSlice := make(map[string][]string)
	shardString := make(map[string]string)
	rbtree := RBTree{}
	s := ShardList{
		ShardSlice:  shardSlice,
		ShardString: shardString,
		Tree:        rbtree,
		NumShards:   numShards,
		PrimaryIP:   primaryIP,
	}

	// take the view and split it into individual server IPs
	sp := strings.Split(globalView, ",")

	// sort them
	sort.Strings(sp)

	// take our list of shard names and sort it
	sort.Strings(shardNames)

	// iterate over the servers
	for i := 0; i < len(sp); i++ {
		// index them into the map, mod the number of shards
		shardIndex := i % numShards

		// the shard id is the index into the name list
		name := shardNames[shardIndex]

		// append them to the list
		s.ShardSlice[name] = append(s.ShardSlice[name], sp[i])

		// check if this particular server is us and assign our shard id
		if sp[i] == s.PrimaryIP {
			s.PrimaryShard = name
		}
	}

	// now insert them into the joined version
	for k, v := range s.ShardSlice {
		s.ShardString[k] = strings.Join(v, ",")
	}

	// build the red black tree
	var tree RBTree

	for k := range s.ShardSlice {
		for _, i := range getVirtualNodePositions(k) {
			tree.put(i, k)
		}
	}

	return &s
}

// ChangeShardNumber is called by the REST API
// Returns true if the change is legal, false otherwise
func (s *ShardList) ChangeShardNumber(n int) bool {
	if s.Size/n < 2 {
		return false
	}

	// Get our list of servers
	str := s.String()
	sl := strings.Split(str, ",")
	sort.Strings(sl)

	// We'll make a new map for them
	newMap := make(map[string][]string)

	// iterate over the servers
	for i := 0; i < len(sl); i++ {
		// index them into the map, mod the number of shards
		shardIndex := i % s.NumShards

		// the shard id is the index into the name list
		name := shardNames[shardIndex]

		// append them to the list
		newMap[name] = append(newMap[name], sl[i])
	}

	// make a shardglob
	sg := ShardGlob{ShardList: newMap}

	s.Overwrite(sg)

	return true
}

// ShuffleServers redistributes servers among shards
func (s *ShardList) ShuffleServers() {

}
