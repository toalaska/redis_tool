package main

import (
	"flag"
	"fmt"
	"gopkg.in/redis.v5"
	"log"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

var (
	REDIS_HOST = ""
	REDIS_PORT = 0
	REDIS_AUTH = ""
	REDIS_DB   = 0

	PRIFIX    = ""
	FILE_NAME = ""

	REDIS_CLIENT *redis.Client
)

func get_keys() []string {
	var pos uint64 = 0
	keys_nums := 0
	all_keys := []string{}

	for true {

		ret := REDIS_CLIENT.Scan(pos, PRIFIX, 10000)
		if ret.Err() != nil {
			log.Fatal("scan err", ret.Err())
		}
		keys, _pos := ret.Val()
		keys_nums += len(keys)
		pos = _pos

		all_keys = append(all_keys, keys...)
		if pos == 0 {
			break;
		}
	}

	return all_keys
}

type String struct {
	Key string `json:"key"`
	Val string `json:"val"`
}
type HashData struct {
	Field string `json:"field"`
	Val   string `json:"val"`
}
type Hash struct {
	Key  string     `json:"key"`
	Data []HashData `json:"data"`
}
type ZsetData struct {
	Member string  `json:"member"`
	Score  float64 `json:"score"`
}
type Zset struct {
	Key  string      `json:"key"`
	Data [] ZsetData `json:"data"`
}
type Set struct {
	Key  string   `json:"key"`
	Data []string `json:"data"`
}
type List struct {
	Key  string   `json:"key"`
	Data []string `json:"data"`
}

type RedisInfo struct {
	String []String `json:"string"`
	Hash   []Hash   `json:"hash"`
	Set    []Set    `json:"set"`
	List   []List   `json:"list"`
	Zset   []Zset   `json:"zset"`
}

func main() {

	if len(os.Args) <= 1  {
		log.Fatal("请输入子命令: dump|import|clear")
	}
	flag.StringVar(&REDIS_HOST, "h", "127.0.0.1", "redis host")
	flag.IntVar(&REDIS_PORT, "p", 6379, "redis port")
	flag.StringVar(&REDIS_AUTH, "a", "", "redis passpord")
	flag.IntVar(&REDIS_DB, "d", 0, "redis db")

	flag.StringVar(&PRIFIX, "P", "", "redis key prefix")
	flag.StringVar(&FILE_NAME, "f", "", "导入或导出数据的文件名")

	flag.CommandLine.Parse(os.Args[2:])

	if os.Args[1]== "-h"{
		flag.PrintDefaults()
		return
	}
	if strings.HasPrefix(os.Args[1], "-"){
				log.Fatal("请输入子命令: dump|import|clear")

	}
	sub_cmd := os.Args[1]

	//
	REDIS_CLIENT = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", REDIS_HOST, REDIS_PORT),
		Password: REDIS_AUTH,
		DB:       REDIS_DB,
	})
	_, err := REDIS_CLIENT.Ping().Result()
	if err != nil {
		log.Fatal("连接redis失败", err)
	}

	switch sub_cmd {
	case "dump":
		chk_prefix()
		chk_file()
		handle_export()
	case "import":
		chk_file()
		handle_import()

	case "clear":
		chk_prefix()
		handle_clear()
	default:
		log.Fatal("非法子命令")
	}

}
func chk_prefix() {
	if PRIFIX == "" {
		log.Fatal("请使用-P参数指定key的前缀")
	}
}
func chk_file() {
	if FILE_NAME == "" {
		log.Fatal("请使用-f参数指定导入导出的文件名")
	}
}
func handle_clear() {
	keys := get_keys()

	ret := REDIS_CLIENT.Del(keys...)
	fmt.Println("成功删除:", ret.Val())
	os.Exit(0)
}
func handle_export() {
	redis_info := RedisInfo{}
	keys := get_keys()
	for _, key := range keys {
		log.Println("dump", key)

		ret := REDIS_CLIENT.Type(key)
		switch ret.Val() {
		case "string":
			ret := REDIS_CLIENT.Get(key)
			val := ret.Val()
			str_info := String{
				Key: key,
				Val: val,
			}
			redis_info.String = append(redis_info.String, str_info)
		case "hash":
			ret := REDIS_CLIENT.HGetAll(key)
			hash_info := Hash{
				Key: key,
			}

			for k, v := range ret.Val() {
				data := HashData{
					Field: k,
					Val:   v,
				}
				hash_info.Data = append(hash_info.Data, data)
			}
			redis_info.Hash = append(redis_info.Hash, hash_info)
		case "set":
			ret := REDIS_CLIENT.SMembers(key)
			set_info := Set{
				Key:  key,
				Data: ret.Val(),
			}
			redis_info.Set = append(redis_info.Set, set_info)

		case "zset":
			ret := REDIS_CLIENT.ZRangeWithScores(key, 0, -1)
			zset_info := Zset{
				Key: key,
			}
			zs := []ZsetData{}
			for _, z := range ret.Val() {
				zs = append(zs, ZsetData{
					Member: z.Member.(string),
					Score:  z.Score,
				})
			}
			zset_info.Data = zs
			redis_info.Zset = append(redis_info.Zset, zset_info)
		case "list":
			ret := REDIS_CLIENT.LRange(key, 0, -1)
			list_info := List{
				Key:  key,
				Data: ret.Val(),
			}
			redis_info.List = append(redis_info.List, list_info)
		default:
			fmt.Println("暂未支持", ret.Val())
		}
	}
	bs, err := json.Marshal(redis_info)
	if err != nil {
		log.Fatal("序列化成json失败", err)
	}
	err = ioutil.WriteFile(FILE_NAME, bs, 0644)
	if err != nil {
		log.Fatal("保存到文件失败", err)
	}
}
func handle_import() {

	bs, err := ioutil.ReadFile(FILE_NAME)
	if err != nil {
		log.Fatal("读取文件失败", err)
	}
	redis_info := RedisInfo{}
	err = json.Unmarshal(bs, &redis_info)
	if err != nil {
		log.Fatal("不是合法的json文件", err)
	}
	//string
	for _, v := range redis_info.String {
		log.Println("import", v.Key)
		REDIS_CLIENT.Set(v.Key, v.Val, 0)
	}
	//hash
	for _, v := range redis_info.Hash {
		log.Println("import", v.Key)

		maps := map[string]string{}
		for _, d := range v.Data {
			maps[d.Field] = d.Val
		}
		REDIS_CLIENT.HMSet(v.Key, maps)
	}
	//set
	for _, v := range redis_info.Set {
		log.Println("import", v.Key)

		for _, v1 := range v.Data {
			REDIS_CLIENT.SAdd(v.Key, v1)
		}
	}
	//zset

	for _, v := range redis_info.Zset {
		log.Println("import", v.Key)

		zs := []redis.Z{}
		for _, v1 := range v.Data {
			zs = append(zs, redis.Z{
				Member: v1.Member,
				Score:  v1.Score,
			})
		}

		REDIS_CLIENT.ZAdd(v.Key, zs...)
	}
	//list
	for _, v := range redis_info.List {
		log.Println("import", v.Key)
		for _, v1 := range v.Data {
			REDIS_CLIENT.RPush(v.Key, v1)
		}
	}

}
