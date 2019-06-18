package p3

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nvgruel123/cs686-blockchain-p3-nvgruel123/p1"
	"github.com/nvgruel123/cs686-blockchain-p3-nvgruel123/p2"
	"github.com/nvgruel123/cs686-blockchain-p3-nvgruel123/p3/data"
)

type Message struct {
	HeartBeatJson []byte
	Signature     []byte
}

var TA_SERVER = "http://localhost:6688"
var REGISTER_SERVER = TA_SERVER + "/peer"
var BC_DOWNLOAD_SERVER = FIRST_NODE_ADDRESS + "/upload"
var SELF_PORT = "3050"
var SELF_ADDR = "http://localhost:"
var FIRST_NODE_ADDRESS = "http://localhost:3050"

var SBC data.SyncBlockChain
var Peers data.PeerList
var ifStarted bool
var PrivateKey []byte
var SelfID []byte

// This function will be executed before everything else.
// Do some initialization here.
func init() {

	SBC = data.NewBlockChain()
	block := p2.Block{}
	mpt := p1.MerklePatriciaTrie{}
	mpt.Initial()
	block.Initial(0, "genesis block", mpt)
	User := make(map[int32][]byte)
	file, _ := os.Open("./rsa/user1_public.pem")
	defer file.Close()
	pub, _ := ioutil.ReadAll(file)
	User[3050] = pub
	file, _ = os.Open("./rsa/user2_public.pem")
	defer file.Close()
	pub, _ = ioutil.ReadAll(file)
	User[3051] = pub
	file, _ = os.Open("./rsa/user3_public.pem")
	defer file.Close()
	pub, _ = ioutil.ReadAll(file)
	User[3052] = pub
	block.User = User
	SBC.Insert(block)
	Peers = data.NewPeerList(0, 32)
	ifStarted = false

}

// Register ID, download BlockChain, start HeartBeat
func Start(w http.ResponseWriter, r *http.Request) {

	if ifStarted == false {
		Register()
		Download()
		ifStarted = true
		fmt.Fprintf(w, "started")
	}
}

// Display peerList and sbc
func Show(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())

}

func SetSelfPort(port string) {

	SELF_PORT = port
	SELF_ADDR = SELF_ADDR + SELF_PORT

}

// Register to TA's server, get an ID
func Register() {

	id, _ := strconv.ParseInt(SELF_PORT, 10, 32)
	Peers.SelfId = int32(id)

}

// send upload request
func Download() {
	if SELF_ADDR != FIRST_NODE_ADDRESS {
		selfId := Peers.SelfId
		url := fmt.Sprintf("%s?id=%v&ip=%s", BC_DOWNLOAD_SERVER, selfId, SELF_ADDR)
		resp, _ := http.Get(url)
		body, _ := ioutil.ReadAll(resp.Body)
		SBC.UpdateEntireBlockChain(string(body))
		Peers.Add("http://localhost:3050", int32(3050))
	}
}

// Upload blockchain to whoever called this method, return jsonStr
func Upload(w http.ResponseWriter, r *http.Request) {

	blockChainJson, _ := SBC.BlockChainToJson()
	u := r.URL
	m, _ := url.ParseQuery(u.RawQuery)
	id, _ := strconv.ParseInt(m["id"][0], 10, 32)
	ip := m["ip"][0]
	Peers.Add(ip, int32(id))
	fmt.Fprint(w, blockChainJson)

}

// Upload a block to whoever called this method, return jsonStr
func UploadBlock(w http.ResponseWriter, r *http.Request) {

	p := strings.Split(r.URL.Path, "/")
	height, err := strconv.ParseInt(p[2], 10, 32)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "InternalServerError", 500)
	}
	hash := p[3]
	block, boolean := SBC.GetBlock(int32(height), hash)
	if boolean == true {
		fmt.Fprint(w, block.EncodeToJson())
	} else {
		http.Error(w, "StatusNoContent", 204)
	}

}

// Received a heartbeat
func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {

	var message Message
	var data data.HeartBeatData
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&message)
	json.Unmarshal(message.HeartBeatJson, &data)
	sender_id := data.Id
	signature := message.Signature
	user := SBC.GetFirstBlock().User
	publicKey := user[sender_id]
	err := RsaSignVer([]byte(message.HeartBeatJson), signature, publicKey)
	if err == nil {
		Peers.Add(data.Addr, data.Id)
		Peers.InjectPeerMapJson(data.PeerMapJson, SELF_ADDR)
		if data.Addr != SELF_ADDR {
			new_block := p2.DecodeFromJson(data.BlockJson)
			timestamp := new_block.Header.Timestamp // block create time
			time_caught, _ := strconv.ParseInt(new_block.Value.Get("time_caught"), 10, 64) // fish caught time

			timeNow := int64(time.Now().Unix()) // current time receive heartbeat
			if time_caught <= timestamp && (timeNow-timestamp) <= 1800 {
				// accept this block
				if !SBC.CheckParentHash(new_block) {
					parentHeight := new_block.Header.Height - 1
					parentHash := new_block.Header.ParentHash
					AskForBlock(parentHeight, parentHash)
				}
				SBC.Insert(new_block)
			}
		}
	}

	data.Hops = data.Hops - 1
	dataJSON, _ := json.Marshal(&data)
	message.HeartBeatJson = dataJSON
	if data.Hops > 0 {
		ForwardHeartBeat(message)
	}

}

// Ask another server to return a block of certain height and hash
func AskForBlock(height int32, hash string) {

	Peers.Rebalance()
	PeerMap := Peers.GetPeerMap()
	_, boolean := SBC.GetBlock(height, hash)
	if boolean == false {
		for ip := range PeerMap {
			url := fmt.Sprintf("%s/block/%v/%s", ip, height, hash)
			resp, _ := http.Get(url)
			if resp.StatusCode == 200 {
				body, _ := ioutil.ReadAll(resp.Body)
				block := p2.DecodeFromJson(string(body))
				parentHeight := height - 1
				parentHash := block.Header.ParentHash
				AskForBlock(parentHeight, parentHash)
				SBC.Insert(block)
				break
			}
		}
	}

}

func ForwardHeartBeat(message Message) {

	Peers.Rebalance()
	PeerMap := Peers.GetPeerMap()
	MessageJson, _ := json.Marshal(message)
	for ip := range PeerMap {
		url := fmt.Sprintf("%s/heartbeat/receive", ip)
		http.Post(url, "application/json; charset=UTF-8", strings.NewReader(string(MessageJson)))
	}

}

// create a block, send to all peers
func CreateBlock(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)
	// get formdata
	r.ParseForm()
	formData := r.Form
	Type := formData.Get("type")
	Count, _ := strconv.Atoi(formData.Get("count"))
	Weight, _ := strconv.ParseFloat(formData.Get("weight"), 64)
	Timecaught := formData.Get("time_caught")
	format := "2006-01-02"
	t, _ := time.Parse(format, Timecaught)
	t64 := int64(t.Unix())
	// set new block
	newBlock := p2.Block{}
	latestBlocks := SBC.GetLatestBlocks()
	latestBlock := latestBlocks[rand.Intn(len(latestBlocks))]
	parentHash := latestBlock.Header.Hash
	parentHeight := latestBlock.Header.Height
	mpt := p1.MerklePatriciaTrie{}
	mpt.Initial()
	mpt.Insert("type", Type)
	mpt.Insert("count", strconv.Itoa(Count))
	mpt.Insert("weight", fmt.Sprintf("%v", Weight))
	mpt.Insert("time_caught", fmt.Sprintf("%v", t64))
	newBlock.Initial(parentHeight+1, parentHash, mpt)
	SBC.Insert(newBlock)
	blockJson := newBlock.EncodeToJson()
	PeerMapJSON, _ := Peers.PeerMapToJson()
	data := data.NewHeartBeatData(true, Peers.SelfId, blockJson, PeerMapJSON, SELF_ADDR)
	heartBeatJson, _ := json.Marshal(data)

	// create signature using private key
	signed, _ := RsaSign([]byte(heartBeatJson), PrivateKey)
	var message Message
	message.HeartBeatJson = heartBeatJson
	message.Signature = signed
	messageJson, _ := json.Marshal(message)
	Peers.Rebalance()
	PeerMap := Peers.GetPeerMap()
	for ip := range PeerMap {
		uri := fmt.Sprintf("%s/heartbeat/receive", ip)
		http.Post(uri, "application/json; charset=UTF-8", strings.NewReader(string(messageJson)))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(messageJson))
}

func SetPrivateKey(priv_path string, pub_path string) {
	priv, _ := ioutil.ReadFile(priv_path)
	pub, _ := ioutil.ReadFile(pub_path)
	PrivateKey = priv
	SelfID = pub
}

func DisplayBlock(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	r.ParseForm()
	formData := r.Form
	height64, _ := strconv.ParseInt(formData.Get("height"), 10, 32)
	height := int32(height64)
	hash := formData.Get("hash")
	block, _ := SBC.GetBlock(height, hash)
	blockJSON := block.EncodeToJson()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(blockJSON))
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
