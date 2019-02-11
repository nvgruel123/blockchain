package main

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"

	"golang.org/x/crypto/sha3"
)

type Flag_value struct {
	encoded_prefix []uint8 // prefix + key end  0, 1 ext 2, 3 leaf
	value          string  // value in leaf, hash value of next node in ext
}

type Node struct {
	node_type    int        // 0: Null, 1: Branch, 2: Ext or Leaf
	branch_value [17]string //not in ext or leaf
	flag_value   Flag_value // not in branch
}

type MerklePatriciaTrie struct {
	db   map[string]Node // hash value of the node
	root string          // hash value
}

func (mpt *MerklePatriciaTrie) Get(key string) string {
	// TODO

	root := mpt.db[mpt.root]
	key_hex := str2hex(key)
	result := ""
	if key == "" {
		return ""
	}
	if root.node_type == 1 {
		// root is branch
		branch_value := root.branch_value
		index := getIndex(key_hex)
		next_node_hash := branch_value[index]
		new_trie := MerklePatriciaTrie{mpt.db, next_node_hash}
		result = new_trie.Get_hex(key_hex[1:])
	} else if root.node_type == 2 {
		// root is leaf or ext
		flag_value := root.flag_value
		encoded_prefix := flag_value.encoded_prefix
		prefix := getPrefix(encoded_prefix)
		if prefix == 0 || prefix == 1 {
			// root is ext
			shared_nibbles := compact_decode(encoded_prefix)
			nibbles_str := array2str(shared_nibbles)
			length := len(nibbles_str)
			if key_hex[:length] == nibbles_str {
				new_key := key_hex[length+1:]
				next_node_hash := root.flag_value.value
				new_trie := MerklePatriciaTrie{mpt.db, next_node_hash}
				result = new_trie.Get_hex(new_key)
			}
		} else {
			// root is leaf

			shared_nibbles := compact_decode(encoded_prefix)
			nibbles_str := array2str(shared_nibbles)
			if nibbles_str == key_hex {
				result = root.flag_value.value
			}
		}
	}
	return result
}

func (mpt *MerklePatriciaTrie) Get_hex(key_hex string) string {
	root := mpt.db[mpt.root]
	result := ""
	if key_hex == "" {
		// base case
		if root.node_type == 1 {
			// end at branch node
			result = root.branch_value[16]
		} else if root.node_type == 2 {
			encoded_prefix := root.flag_value.encoded_prefix
			prefix := encoded_prefix[0] / 16
			if prefix == 2 || prefix == 3 {
				// leaf
				result = root.flag_value.value
			}
		}
	} else {
		// recursive step
		if root.node_type == 1 {
			// branch node
			branch_value := root.branch_value
			index := getIndex(key_hex)
			next_node_hash := branch_value[index]
			new_trie := MerklePatriciaTrie{mpt.db, next_node_hash}
			result = new_trie.Get_hex(key_hex[1:])
		} else {
			encoded_prefix := root.flag_value.encoded_prefix
			prefix := encoded_prefix[0] / 16
			if prefix == 0 || prefix == 1 {
				// ext
				shared_nibbles := compact_decode(encoded_prefix)
				nibbles_str := array2str(shared_nibbles)
				length := len(nibbles_str)
				if key_hex[:length] == nibbles_str {
					new_key := key_hex[length+1:]
					next_node_hash := root.flag_value.value
					new_trie := MerklePatriciaTrie{mpt.db, next_node_hash}
					result = new_trie.Get_hex(new_key)
				}
			} else if prefix == 2 || prefix == 3 {
				// leaf
				shared_nibbles := compact_decode(encoded_prefix)
				nibbles_str := array2str(shared_nibbles)
				if nibbles_str == key_hex {
					result = root.flag_value.value
				}
			}
		}
	}
	return result
}

func array2str(array []uint8) string {
	result := ""
	for _, v := range array {
		result += strconv.Itoa(int(v))
	}
	return result
}

func hex2array(key_hex string) []uint8 {
	array := []uint8{}
	for _, c := range key_hex {
		chr := string(c)
		int_value, _ := strconv.Atoi(chr)
		uint_value := uint8(int_value)
		array = append(array, uint_value)
	}
	return array
}

func getPrefix(encoded_prefix []uint8) uint8 {
	result := encoded_prefix[0] / 16
	return result
}

func (mpt *MerklePatriciaTrie) Insert(key string, new_value string) {
	// TODO
	root := mpt.db[mpt.root]
	db := mpt.db
	key_hex := str2hex(key)
	if root.node_type == 0 {
		// empty tree
		hex_array := hex2array(key_hex)
		hex_array = append(hex_array, 0x10)
		encoded_arr := compact_encode(hex_array)
		flag_value := Flag_value{encoded_arr, new_value}
		new_leaf_node := Node{2, [17]string{}, flag_value}
		db[new_leaf_node.hash_node()] = new_leaf_node
		mpt.root = new_leaf_node.hash_node()
	} else if root.node_type == 1 {
		// branch
		branch_value := root.branch_value
		index := getIndex(key_hex)
		new_key_hex := key_hex[1:]
		if new_key_hex == "" {
			// insert new_value to index 16, update mpt
			branch_value[16] = new_value
			hashed_branch := root.hash_node()
			db[hashed_branch] = root
			mpt.root = hashed_branch
		} else {
			if branch_value[index] == "" {
				// nothing on that index, create new leaf node
				hex_array := hex2array(new_key_hex[1:])
				hex_array = append(hex_array, 0x10)
				encoded_arr := compact_encode(hex_array)
				flag_value := Flag_value{encoded_arr, new_value}
				new_leaf := Node{2, [17]string{}, flag_value}
				hash_new_leaf := new_leaf.hash_node()

				// connect new leaf to branch and update branch node
				branch_value[index] = hash_new_leaf
				new_root := Node{1, branch_value, Flag_value{}}
				new_hash := mpt.UpdateNode(new_root)
				mpt.root = new_hash
			} else {
				// something on index, use this node as root and call insert on it
				next_node_hash := branch_value[index]
				new_trie := MerklePatriciaTrie{mpt.db, next_node_hash}
				hashed_node := new_trie.Insert_hex(new_key_hex, new_value)

				// connect result to branch and update branch node
				branch_value[index] = hashed_node
				new_root := Node{1, branch_value, Flag_value{}}
				new_hash := mpt.UpdateNode(new_root)
				mpt.root = new_hash
			}
		}
	} else {
		// ext or leaf
		flag_value := root.flag_value
		encoded_prefix := flag_value.encoded_prefix
		prefix := getPrefix(encoded_prefix)
		if prefix == 0 || prefix == 1 {
			// ext
			shared_nibbles := compact_decode(encoded_prefix)
			nibbles_str := array2str(shared_nibbles) //shared nibbles
			common_path := getCommonPath(key_hex, nibbles_str)
			common_length := len(common_path)
			rest_path := key_hex[common_length:]
			rest_nibbles := nibbles_str[common_length:]
			if common_length == 0 {
				new_branch_value := [17]string{}

				// create new leaf node to insert
				index_path := getIndex(rest_path)
				leaf_end := rest_path[1:]
				hex_array := hex2array(leaf_end)
				hex_array = append(hex_array, 0x10)
				encoded_arr := compact_encode(hex_array)
				flag_value := Flag_value{encoded_arr, new_value}
				new_leaf := Node{2, [17]string{}, flag_value}
				hashed_new_leaf := new_leaf.hash_node()
				mpt.db[hashed_new_leaf] = new_leaf
				new_branch_value[index_path] = hashed_new_leaf

				// create new ext node and connect to old path
				index_nibble := getIndex(rest_nibbles)
				new_nibbles := rest_nibbles[1:]
				hex_array = hex2array(new_nibbles)
				encoded_arr = compact_encode(hex_array)
				next_node_hash := root.flag_value.value              // original hash
				flag_value = Flag_value{encoded_arr, next_node_hash} // connect to original trie
				new_ext_node := Node{2, [17]string{}, flag_value}
				hashed_new_ext := new_ext_node.hash_node()
				mpt.db[hashed_new_ext] = new_ext_node
				new_branch_value[index_nibble] = hashed_new_ext
				// create branch node
				new_branch_node := Node{1, new_branch_value, Flag_value{}}
				hashed_new_branch := new_branch_node.hash_node()
				mpt.db[hashed_new_branch] = new_branch_node
				mpt.root = hashed_new_branch
			} else { // common path not 0
				// create ext node for common path
				hashed_new_branch := ""
				if rest_nibbles != "" {
					new_branch_value := [17]string{}

					// create new ext node
					new_branch_node := Node{1, new_branch_value, Flag_value{}}
					hashed_new_branch = new_branch_node.hash_node()
					mpt.db[hashed_new_branch] = new_branch_node
					nibble_index := getIndex(rest_nibbles)
					hex_aray := hex2array(rest_nibbles[1:])
					encoded_arr := compact_encode(hex_aray)
					new_flag_value := Flag_value{encoded_arr, root.flag_value.value} // original value
					new_ext_node := Node{2, [17]string{}, new_flag_value}
					hashed_new_ext := new_ext_node.hash_node()
					mpt.db[hashed_new_ext] = new_ext_node
					new_branch_value[nibble_index] = hashed_new_ext
					new_trie := MerklePatriciaTrie{db, hashed_new_branch}
					hashed_new_branch = new_trie.Insert_hex(rest_path, new_value)
				}

			}
		} else if prefix == 2 || prefix == 3 {
			// leaf

		}
	}

}

func getCommonPath(key_hex string, nibbles_str string) string {
	common_path := ""
	for i, _ := range key_hex {
		if key_hex[i] == nibbles_str[i] {
			common_path += fmt.Sprintf("%c", key_hex[i])
		} else {
			break
		}
	}
	return common_path
}

func (mpt *MerklePatriciaTrie) Insert_hex(key_hex string, new_value string) string {
	hashed_node := ""
	// root := mpt.db[mpt.root]
	// db := mpt.db
	return hashed_node
}

func (mpt *MerklePatriciaTrie) UpdateNode(node Node) string {
	hashed_node := node.hash_node()
	mpt.db[hashed_node] = node
	return hashed_node
}

func str2hex(key string) string {
	return fmt.Sprintf("%x", key)
}

func (mpt *MerklePatriciaTrie) Delete(key string) {
	// TODO
}

func compact_encode(hex_array []uint8) []uint8 {
	// TODO
	var term int
	if hex_array[len(hex_array)-1] == 16 {
		term = 1
	} else {
		term = 0
	}
	if term == 1 {
		hex_array = hex_array[:len(hex_array)-1]
	}
	var oddlen = len(hex_array) % 2
	var flags = 2*term + oddlen
	if oddlen == 1 {
		array := []uint8{uint8(flags)}
		for _, v := range hex_array {
			array = append(array, v)
		}
		hex_array = array
	} else {
		array := []uint8{uint8(flags), 0}
		for _, v := range hex_array {
			array = append(array, v)
		}
		hex_array = array
	}
	o := []uint8{}
	for i := 0; i < len(hex_array); i += 2 {
		o = append(o, 16*hex_array[i]+hex_array[i+1])
	}
	return o
}

// If Leaf, ignore 16 at the end
func compact_decode(encoded_arr []uint8) []uint8 {
	// TODO
	hex_array := []uint8{}
	for _, v := range encoded_arr {
		x := v % 16
		y := v / 16
		hex_array = append(hex_array, y)
		hex_array = append(hex_array, x)
	}
	flags := hex_array[0]
	oddlen := flags % 2
	if oddlen == 1 {
		hex_array = hex_array[1:]
	} else {
		hex_array = hex_array[2:]
	}
	return hex_array
}

func test_compact_encode() {
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))
}

func (node *Node) hash_node() string {
	var str string
	switch node.node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.branch_value {
			str += v
		}
	case 2:
		str = node.flag_value.value
	}

	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func getIndex(key_hex string) int {
	index, _ := strconv.Atoi(fmt.Sprintf("%c", key_hex[0]))
	return index
}

func (node *Node) toString() string {
	result := ""
	result += fmt.Sprintf("Node:\n  node_type: %v\n  branch_value: %v\n  flag_value: %v\n", node.node_type, node.branch_value, node.flag_value.toString())
	return result
}

func (flag_value *Flag_value) toString() string {
	result := ""
	array := "["
	for _, v := range flag_value.encoded_prefix {
		array += fmt.Sprintf("%x ", v)
	}
	array = array[:len(array)-1]
	array += "]"
	result += fmt.Sprintf("encoded_prefix: %v\n  value: %v\n", array, flag_value.value)
	return result
}

func (mpt *MerklePatriciaTrie) toString() string {
	result := ""
	db := mpt.db
	for hash, node := range db {
		result += fmt.Sprintf("%v\n  %v\n", hash, node.toString())
	}
	result += fmt.Sprintf("root: %v", mpt.root)
	return result
}

func main() {
	db := make(map[string]Node)
	branch_value := [17]string{}
	slice := []uint8{}
	flag_value := Flag_value{slice, ""}
	root := Node{0, branch_value, flag_value}
	hash_value := root.hash_node()
	db[hash_value] = root
	mpt := MerklePatriciaTrie{db, hash_value}
	mpt.Insert("ab", "apple")
	fmt.Println(mpt.toString())
	branch_value[1] = "adsf"
	root = Node{0, branch_value, flag_value}
	hash := mpt.UpdateNode(root)
	mpt.root = hash
	fmt.Println(mpt.toString())
}
