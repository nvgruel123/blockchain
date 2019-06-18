package data

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

type PeerList struct {
	SelfId    int32            `json:"selfId"`
	PeerMap   map[string]int32 `json:"peerMap"` // string ip, int32 id
	MaxLength int32            `json:"maxLength"`
	Mux       sync.Mutex       `json:"mux"`
}

func NewPeerList(id int32, maxLength int32) PeerList {

	peerMap := make(map[string]int32)
	peerlist := PeerList{}
	peerlist.SelfId = id
	peerlist.PeerMap = peerMap
	peerlist.MaxLength = maxLength

	return peerlist
}

func (peers *PeerList) GetPeerMap() map[string]int32 {
	return peers.Copy()
}

func (peers *PeerList) Add(addr string, id int32) {

	peers.Mux.Lock()
	if peers.SelfId != id {
		peerMap := peers.PeerMap
		peerMap[addr] = id
	}
	peers.Mux.Unlock()

}

func (peers *PeerList) Delete(addr string) {

	peers.Mux.Lock()
	peerMap := peers.PeerMap
	delete(peerMap, addr)
	peers.Mux.Unlock()

}

func (peers *PeerList) Rebalance() {

	peers.Mux.Lock()
	peerMap := peers.GetPeerMap()
	length := len(peerMap)
	maxLength := peers.MaxLength
	if length > int(maxLength) {
		selfID := peers.SelfId
		var ids []int
		for _, v := range peerMap {
			ids = append(ids, int(v))
		}
		ids = append(ids, int(selfID))
		sort.Ints(ids)
		index := indexOf(int(selfID), ids)
		newPeerMap := make(map[string]int32)
		for i := 1; i <= int(maxLength)/2; i++ {
			idx := index + i
			if idx <= length {
				ip := getKey(peerMap, ids[idx])
				newPeerMap[ip] = int32(ids[idx])
			} else {
				ip := getKey(peerMap, ids[idx-length-1])
				newPeerMap[ip] = int32(ids[idx-length-1])
			}
			idx = index - i
			if idx >= 0 {
				ip := getKey(peerMap, ids[idx])
				newPeerMap[ip] = int32(ids[idx])
			} else {
				ip := getKey(peerMap, ids[idx+length+1])
				newPeerMap[ip] = int32(ids[idx+length+1])
			}
		}
		peers.PeerMap = newPeerMap
	}
	peers.Mux.Unlock()

}

func getKey(peerMap map[string]int32, id int) string {

	res := ""
	for k, v := range peerMap {
		if int(v) == id {
			res = k
		}
	}

	return res
}

func indexOf(selfID int, ids []int) int {

	for k, v := range ids {
		if selfID == v {
			return k
		}
	}

	return -1
}

func (peers *PeerList) Show() string {

	rs := ""
	for ip, id := range peers.PeerMap {
		rs += fmt.Sprintf("%s: ", ip)
		rs += fmt.Sprintf("%v, ", int(id))
	}
	rs = fmt.Sprintf("This is the PeerList: %s", rs)
	return rs
}

func (peers *PeerList) Register(id int32) {

	peers.Mux.Lock()
	peers.SelfId = id
	peers.Mux.Unlock()

}

func (peers *PeerList) Copy() map[string]int32 {

	copy := make(map[string]int32)
	peerMap := peers.PeerMap
	for ip, id := range peerMap {
		copy[ip] = id
	}

	return copy
}

func (peers *PeerList) GetSelfId() int32 {
	return peers.SelfId
}

func (peers *PeerList) PeerMapToJson() (string, error) {

	byteArr, err := json.Marshal(peers.PeerMap)
	return string(byteArr), err

}

func (peers *PeerList) InjectPeerMapJson(peerMapJsonStr string, selfAddr string) {

	peers.Mux.Lock()
	byteArr := []byte(peerMapJsonStr)
	var peerMap map[string]int32
	json.Unmarshal(byteArr, &peerMap)
	for k, v := range peerMap {
		if k != selfAddr {
			peers.PeerMap[k] = v
		}
	}
	peers.Mux.Unlock()

}

func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	expected.Add("-1-1", -1)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("7777", 7)
	peers.Add("9999", 9)
	peers.Add("11111111", 11)
	peers.Add("2020", 20)
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("7777", 7)
	expected.Add("9999", 9)
	expected.Add("2020", 20)
	fmt.Println(reflect.DeepEqual(peers, expected))
}
