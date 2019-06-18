package data

type HeartBeatData struct {
	IfNewBlock  bool   `json:"ifNewBlock"`
	Id          int32  `json:"id"`
	BlockJson   string `json:"blockJson"`
	PeerMapJson string `json:"peerMapJson"`
	Addr        string `json:"addr"`
	Hops        int32  `json:"hops"`
}

func NewHeartBeatData(ifNewBlock bool, id int32, blockJson string, peerMapJson string, addr string) HeartBeatData {

	data := HeartBeatData{ifNewBlock, id, blockJson, peerMapJson, addr, 3}
	return data
}

func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapBase64 string, addr string) HeartBeatData {

	data := HeartBeatData{}
	data.IfNewBlock = false
	data.Id = selfId
	data.BlockJson = ""
	data.PeerMapJson = peerMapBase64
	data.Addr = addr
	data.Hops = 3

	return data
}
