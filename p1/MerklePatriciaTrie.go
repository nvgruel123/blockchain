package p1

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
)

type Flag_value struct {
	Encoded_prefix []uint8 // prefix + key end  0, 1 ext 2, 3 leaf
	Value          string  // value in leaf, hash value of next node in ext
}

type Node struct {
	Node_type    int        // 0: Null, 1: Branch, 2: Ext or Leaf
	Branch_value [17]string //not in ext or leaf
	Flag_value   Flag_value // not in branch
}

type MerklePatriciaTrie struct {
	Db   map[string]Node   // hash value of the node
	Root string            // hash value
	Data map[string]string // key value pairs
}

func (mpt *MerklePatriciaTrie) Get(key string) string {
	// TODO

	root := mpt.Db[mpt.Root]
	key_hex := str2hex(key)
	result := ""

	if root.Node_type != 0 {
		result = mpt.Get_Rec(key_hex)
	}
	return result
}

func (mpt *MerklePatriciaTrie) Get_Rec(key_hex string) string {
	root := mpt.Db[mpt.Root]
	db := mpt.Db
	data := mpt.Data
	result := ""
	if root.Node_type == 0 {
		return ""
	} else if root.Node_type == 1 {
		// branch
		if key_hex == "" {
			return root.Branch_value[16]
		} else {
			branch_value := root.Branch_value
			key_index := getIndex(key_hex)
			next_node_hash := branch_value[key_index]
			new_trie := MerklePatriciaTrie{db, next_node_hash, data}
			result = new_trie.Get_Rec(key_hex[1:])
		}
	} else {
		// ext or leaf
		encoded_prefix := root.Flag_value.Encoded_prefix
		node_value := root.Flag_value.Value
		decoded_arr := compact_decode(encoded_prefix)
		prefix := getPrefix(encoded_prefix)
		if prefix == 0 || prefix == 1 {
			// ext
			nibbles_str := array2str(decoded_arr)
			if len(nibbles_str) == len(key_hex) {
				if nibbles_str != key_hex {
					return ""
				}
			}
			if getCommonPath(nibbles_str, key_hex) == "" {
				return ""
			}
			if len(key_hex) < len(nibbles_str) {
				return ""
			}
			nibbles_len := len(nibbles_str)
			new_trie := MerklePatriciaTrie{db, node_value, data}
			result = new_trie.Get_Rec(key_hex[nibbles_len:])
		} else {
			// leaf
			key_end := array2str(decoded_arr)
			if key_end == key_hex {
				result = node_value
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
	root := mpt.Db[mpt.Root]
	db := mpt.Db
	data := mpt.Data
	data[key] = new_value
	key_hex := str2hex(key)
	if root.Node_type == 0 {
		hex_array := hex2array(key_hex)
		hex_array = append(hex_array, 0x10)
		encoded_arr := compact_encode(hex_array)
		flag_value := Flag_value{encoded_arr, new_value}
		new_leaf_node := Node{Node_type: 2, Flag_value: flag_value}
		delete(db, mpt.Root)
		new_leaf_hash := mpt.UpdateNode(new_leaf_node)
		mpt.Root = new_leaf_hash
	} else {
		hashed_node := mpt.Insert_Rec(key_hex, new_value)
		mpt.Root = hashed_node
	}
}

func getCommonPath(key_hex string, nibbles_str string) string {
	common_path := ""

	if len(key_hex) < len(nibbles_str) {
		for i, _ := range key_hex {
			if key_hex[i] == nibbles_str[i] {
				common_path += fmt.Sprintf("%c", key_hex[i])
			} else {
				break
			}
		}
	} else {
		for i, _ := range nibbles_str {
			if key_hex[i] == nibbles_str[i] {
				common_path += fmt.Sprintf("%c", key_hex[i])
			} else {
				break
			}
		}
	}
	return common_path
}

func (mpt *MerklePatriciaTrie) Insert_Rec(key_hex string, new_value string) string {
	hashed_node := ""
	db := mpt.Db
	root := db[mpt.Root]
	data := mpt.Data

	if root.Node_type == 1 {
		// branch node
		branch_value := root.Branch_value

		if key_hex == "" {
			branch_value[16] = new_value
		} else {
			key_index := getIndex(key_hex)
			next_node_hash := branch_value[key_index]
			if next_node_hash == "" {
				key_arr := hex2array(key_hex[1:])
				key_arr = append(key_arr, 0x10)
				encoded_arr := compact_encode(key_arr)
				new_flag_value := Flag_value{encoded_arr, new_value}
				new_leaf := Node{2, [17]string{}, new_flag_value}
				branch_value[key_index] = mpt.UpdateNode(new_leaf)
			} else {

				new_trie := MerklePatriciaTrie{db, next_node_hash, data}
				branch_value[key_index] = new_trie.Insert_Rec(key_hex[1:], new_value)
			}
		}
		new_branch_node := Node{1, branch_value, Flag_value{}}
		hashed_node = mpt.UpdateNode(new_branch_node)
	} else if root.Node_type == 2 {
		encoded_prefix := root.Flag_value.Encoded_prefix
		node_value := root.Flag_value.Value
		decoded_arr := compact_decode(encoded_prefix)
		decoded_str := array2str(decoded_arr)
		prefix := getPrefix(encoded_prefix)
		common_path := getCommonPath(key_hex, decoded_str)
		common_length := len(common_path)
		rest_path := key_hex[common_length:]
		rest_nibbles := decoded_str[common_length:]
		if prefix == 0 || prefix == 1 {
			// ext

			if common_path == decoded_str {
				new_trie := MerklePatriciaTrie{db, node_value, data}
				new_branch_hash := new_trie.Insert_Rec(rest_path, new_value)
				new_flag_value := Flag_value{encoded_prefix, new_branch_hash}
				new_ext_node := Node{2, [17]string{}, new_flag_value}
				hashed_node = mpt.UpdateNode(new_ext_node)
			} else if common_length == 0 {
				// branch node
				new_branch_value := [17]string{}
				if rest_nibbles == "" {
					branch_node := db[node_value]
					branch_value := branch_node.Branch_value
					for index, hash := range branch_value {
						new_branch_value[index] = hash
					}
				} else {
					nibble_index := getIndex(rest_nibbles)
					if len(rest_nibbles) == 1 {
						new_branch_value[nibble_index] = node_value
					} else {
						// optional ext
						optional_ext_nibbles := rest_nibbles[1:]
						optional_ext_arr := hex2array(optional_ext_nibbles)
						optional_encoded_arr := compact_encode(optional_ext_arr)
						optional_flag_value := Flag_value{optional_encoded_arr, node_value}
						optional_ext_node := Node{2, [17]string{}, optional_flag_value}
						optional_ext_hash := mpt.UpdateNode(optional_ext_node)
						new_branch_value[nibble_index] = optional_ext_hash
					}
				}
				new_branch := Node{1, new_branch_value, Flag_value{}}
				new_branch_hash := mpt.UpdateNode(new_branch)
				new_trie := MerklePatriciaTrie{db, new_branch_hash, data}
				hashed_node = new_trie.Insert_Rec(rest_path, new_value)

			} else {
				// ext + branch + ext(optional)
				new_branch_value := [17]string{}
				nibble_index := getIndex(rest_nibbles)
				if rest_nibbles[1:] != "" {
					// another ext
					optional_ext_nibbles := rest_nibbles[1:]
					optional_encoded_arr := compact_encode(hex2array(optional_ext_nibbles))
					optional_flag_value := Flag_value{optional_encoded_arr, node_value}
					optional_ext_node := Node{2, [17]string{}, optional_flag_value}
					optional_ext_hash := mpt.UpdateNode(optional_ext_node)
					new_branch_value[nibble_index] = optional_ext_hash
				} else {
					new_branch_value[nibble_index] = node_value
				}
				new_branch_node := Node{1, new_branch_value, Flag_value{}}
				new_branch_hash := mpt.UpdateNode(new_branch_node)
				new_trie := MerklePatriciaTrie{db, new_branch_hash, data}
				new_branch_hash = new_trie.Insert_Rec(rest_path, new_value)
				new_encoded_arr := compact_encode(hex2array(common_path))
				new_flag_value := Flag_value{new_encoded_arr, new_branch_hash}
				new_ext_node := Node{2, [17]string{}, new_flag_value}
				hashed_node = mpt.UpdateNode(new_ext_node)
			}
		} else {
			// leaf
			if key_hex == decoded_str {
				// same key, update value
				new_hex_arr := hex2array(key_hex)
				new_hex_arr = append(new_hex_arr, 0x10)
				new_encoded_arr := compact_encode(new_hex_arr)
				new_flag_value := Flag_value{new_encoded_arr, new_value}
				new_leaf_node := Node{2, [17]string{}, new_flag_value}
				hashed_node = mpt.UpdateNode(new_leaf_node)
			} else {
				// create new branch
				new_branch_value := [17]string{}
				new_branch := Node{1, new_branch_value, Flag_value{}}
				new_branch_hash := mpt.UpdateNode(new_branch)
				new_trie := MerklePatriciaTrie{db, new_branch_hash, data}
				if common_path == "" {
					new_branch_hash = new_trie.Insert_Rec(key_hex, new_value) // insert new value to empty branch
					new_trie = MerklePatriciaTrie{db, new_branch_hash, data}
					hashed_node = new_trie.Insert_Rec(decoded_str, node_value) // insert old leaf to branch
				} else {
					// create new ext and new branch
					new_branch_hash = new_trie.Insert_Rec(rest_path, new_value)
					new_trie = MerklePatriciaTrie{db, new_branch_hash, data}
					new_branch_hash = new_trie.Insert_Rec(rest_nibbles, node_value)
					new_common_arr := hex2array(common_path)
					new_encoded_arr := compact_encode(new_common_arr)
					new_flag_value := Flag_value{new_encoded_arr, new_branch_hash}
					new_ext_node := Node{2, [17]string{}, new_flag_value}
					hashed_node = mpt.UpdateNode(new_ext_node)
				}

			}
		}
	}
	return hashed_node
}

func (mpt *MerklePatriciaTrie) UpdateNode(node Node) string {
	hashed_node := node.hash_node()
	mpt.Db[hashed_node] = node
	return hashed_node
}

func str2hex(key string) string {
	return fmt.Sprintf("%x", key)
}

func (mpt *MerklePatriciaTrie) Delete(key string) {
	// TODO

	if mpt.Get(key) != "" {
		delete(mpt.Data, key)
		key_hex := str2hex(key)
		mpt.Root = mpt.Delete_Rec(key_hex)
	}

}

func (mpt *MerklePatriciaTrie) Delete_Rec(key_hex string) string {
	hashed_node := ""
	db := mpt.Db
	root := db[mpt.Root]
	data := mpt.Data
	if root.Node_type == 1 {
		// branch
		branch_value := root.Branch_value
		if key_hex == "" {
			// delete value
			branch_value[16] = ""
		} else {
			// delete(update) index
			key_index := getIndex(key_hex)
			next_node_hash := branch_value[key_index]
			new_trie := MerklePatriciaTrie{db, next_node_hash, data}
			next_node_hash = new_trie.Delete_Rec(key_hex[1:])
			branch_value[key_index] = next_node_hash
		}
		// check len of branch array
		branch_arr := getCount(branch_value)
		if len(branch_arr) == 1 {
			index := branch_arr[0]
			if index == 16 {
				// branch -> leaf
				old_value := branch_value[16]
				key_end := ""
				key_arr := hex2array(key_end)
				key_arr = append(key_arr, 0x10)
				encoded_arr := compact_encode(key_arr)
				flag_value := Flag_value{encoded_arr, old_value}
				new_leaf_node := Node{2, [17]string{}, flag_value}
				return mpt.UpdateNode(new_leaf_node)
			} else {
				next_node_hash := branch_value[index]
				next_node := db[next_node_hash]
				if next_node.Node_type == 1 {
					// next node branch, current node to ext
					index_str := fmt.Sprintf("%d", index)
					hex_arr := hex2array(index_str)
					encoded_arr := compact_encode(hex_arr)
					flag_value := Flag_value{encoded_arr, next_node_hash}
					new_ext_node := Node{2, [17]string{}, flag_value}
					return mpt.UpdateNode(new_ext_node)
				} else {
					// next node ext or leaf
					new_key := fmt.Sprintf("%d", index)
					next_node_hash := branch_value[index]
					next_node := db[next_node_hash]
					next_node_encoded_prefix := next_node.Flag_value.Encoded_prefix
					next_node_value := next_node.Flag_value.Value
					prefix := getPrefix(next_node_encoded_prefix)

					if prefix == 0 || prefix == 1 {
						// next node is ext node -> new ext node
						decoded_arr := compact_decode(next_node_encoded_prefix)
						nibbles_str := array2str(decoded_arr)
						new_key += nibbles_str
						new_nibble_arr := hex2array(new_key)
						encoded_arr := compact_encode(new_nibble_arr)
						new_flag_value := Flag_value{encoded_arr, next_node_value}
						new_ext_node := Node{2, [17]string{}, new_flag_value}
						return mpt.UpdateNode(new_ext_node)
					} else {
						// leaf to new leaf
						decoded_arr := compact_decode(next_node_encoded_prefix)
						nibbles_str := array2str(decoded_arr)
						new_key += nibbles_str
						new_nibble_arr := hex2array(new_key)
						new_nibble_arr = append(new_nibble_arr, 0x10)
						encoded_arr := compact_encode(new_nibble_arr)
						new_flag_value := Flag_value{encoded_arr, next_node_value}
						new_leaf_node := Node{2, [17]string{}, new_flag_value}
						return mpt.UpdateNode(new_leaf_node)
					}
				}
			}
		}
		// update branch
		new_branch := Node{1, branch_value, Flag_value{}}
		hashed_node = mpt.UpdateNode(new_branch)
	} else {
		// ext or leaf
		encoded_prefix := root.Flag_value.Encoded_prefix
		prefix := getPrefix(encoded_prefix)
		if prefix == 0 || prefix == 1 {
			// ext
			decoded_arr := compact_decode(encoded_prefix)
			nibble_str := array2str(decoded_arr)
			nibble_len := len(nibble_str)
			new_key_hex := key_hex[nibble_len:]
			next_node_hash := root.Flag_value.Value
			new_trie := MerklePatriciaTrie{db, next_node_hash, data}
			next_node_hash = new_trie.Delete_Rec(new_key_hex)
			next_node := db[next_node_hash]

			if next_node.Node_type == 1 {
				// return node still branch, update current node
				new_flag_value := Flag_value{encoded_prefix, next_node_hash}
				new_ext_node := Node{2, [17]string{}, new_flag_value}
				hashed_node = mpt.UpdateNode(new_ext_node)
			} else {
				// return node not branch
				next_node_value := next_node.Flag_value.Value
				next_node_encoded_prefix := next_node.Flag_value.Encoded_prefix
				next_node_prefix := getPrefix(next_node_encoded_prefix)
				if next_node_prefix == 1 || next_node_prefix == 0 {
					// next node ext node, combine two ext to one
					next_node_decoded_arr := compact_decode(next_node_encoded_prefix)
					next_node_nibble_str := array2str(next_node_decoded_arr)
					new_nibble_str := nibble_str + next_node_nibble_str
					new_nibble_arr := hex2array(new_nibble_str)
					new_encoded_arr := compact_encode(new_nibble_arr)
					new_flag_value := Flag_value{new_encoded_arr, next_node_value}
					new_ext_node := Node{2, [17]string{}, new_flag_value}
					hashed_node = mpt.UpdateNode(new_ext_node)
				} else {
					// next node leaf, combine ext + leaf to leaf
					next_node_decoded_arr := compact_decode(next_node_encoded_prefix)
					next_node_key_end := array2str(next_node_decoded_arr)
					new_key_end := nibble_str + next_node_key_end
					new_key_arr := hex2array(new_key_end)
					new_key_arr = append(new_key_arr, 0x10)
					new_encoded_arr := compact_encode(new_key_arr)
					new_flag_value := Flag_value{new_encoded_arr, next_node_value}
					new_leaf_node := Node{2, [17]string{}, new_flag_value}
					hashed_node = mpt.UpdateNode(new_leaf_node)
				}
			}
		} // current node leaf, do nothing
	}
	return hashed_node
}

func getCount(branch_value [17]string) []int {
	index_arr := []int{}
	for i, value := range branch_value {
		if value != "" {
			index_arr = append(index_arr, i)
		}
	}
	return index_arr
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

func getIndex(key_hex string) int {
	index, _ := strconv.Atoi(fmt.Sprintf("%c", key_hex[0]))
	return index
}
func test_compact_encode() {
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))
}

func (node *Node) hash_node() string {
	var str string
	switch node.Node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.Branch_value {
			str += v
		}
	case 2:
		str = node.Flag_value.Value
	}

	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func (node *Node) String() string {
	str := "empty string"
	switch node.Node_type {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.Branch_value[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("value=%s]", node.Branch_value[16])
	case 2:
		encoded_prefix := node.Flag_value.Encoded_prefix
		node_name := "Leaf"
		if is_ext_node(encoded_prefix) {
			node_name = "Ext"
		}
		ori_prefix := strings.Replace(fmt.Sprint(compact_decode(encoded_prefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, value=\"%s\">", node_name, ori_prefix, node.Flag_value.Value)
	}
	return str
}

func node_to_string(node Node) string {
	return node.String()
}

func (mpt *MerklePatriciaTrie) Initial() {
	mpt.Db = make(map[string]Node)
	mpt.Data = make(map[string]string)
}

func is_ext_node(encoded_arr []uint8) bool {
	return encoded_arr[0]/16 < 2
}

func TestCompact() {
	test_compact_encode()
}

func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.Root)
	for hash := range mpt.Db {
		content += fmt.Sprintf("%s: %s\n", hash, node_to_string(mpt.Db[hash]))
	}
	return content
}

func (mpt *MerklePatriciaTrie) Order_nodes() string {
	raw_content := mpt.String()
	content := strings.Split(raw_content, "\n")
	root_hash := strings.Split(strings.Split(content[0], "HashStart")[1], "HashEnd")[0]
	queue := []string{root_hash}
	i := -1
	rs := ""
	cur_hash := ""
	for len(queue) != 0 {
		last_index := len(queue) - 1
		cur_hash, queue = queue[last_index], queue[:last_index]
		i += 1
		line := ""
		for _, each := range content {
			if strings.HasPrefix(each, "HashStart"+cur_hash+"HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart"+cur_hash+"HashEnd", fmt.Sprintf("Hash%v", i), -1)
			}
		}
		temp2 := strings.Split(line, "HashStart")
		flag := true
		for _, each := range temp2 {
			if flag {
				flag = false
				continue
			}
			queue = append(queue, strings.Split(each, "HashEnd")[0])
		}
	}
	return rs
}
