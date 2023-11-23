package search

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/go-ego/gse"
)

// 分词跳过
var skip = map[string]struct{}{
	" ":  {},
	",":  {},
	".":  {},
	"#":  {},
	"`":  {},
	"\n": {},
	"\r": {},
}

var trim = "\"\n\t `|~!@#$%^&*()_-+=,.<>/?':;；：[]{}\\！，￥…（）—《》。？【】、”“"

// 计算整数切片的交集，并返回按照次数递减排序的列表（不包含重复元素）
func intersection(slices [][]int) []int {
	if len(slices) == 0 {
		return nil
	}

	// 创建一个映射来跟踪元素的出现次数
	elementCount := make(map[int]int)

	// 遍历第一个切片，记录元素出现次数
	for _, element := range slices[0] {
		elementCount[element]++
	}

	// 遍历剩余切片，更新元素出现次数
	for _, slice := range slices[1:] {
		currentElementCount := make(map[int]int)

		// 统计当前切片中元素的出现次数
		for _, element := range slice {
			currentElementCount[element]++
		}

		// 更新elementCount映射，保留最小的出现次数
		for element, count := range elementCount {
			if currentCount, found := currentElementCount[element]; found {
				if currentCount < count {
					elementCount[element] = currentCount
				}
			} else {
				// 如果元素在当前切片中不存在，则将其出现次数设置为0
				elementCount[element] = 0
			}
		}
	}

	// 提取不重复的元素
	var uniqueElements []int
	for element, count := range elementCount {
		if count > 0 {
			uniqueElements = append(uniqueElements, element)
		}
	}

	// 按照次数递减排序
	sort.Slice(uniqueElements, func(i, j int) bool {
		return elementCount[uniqueElements[i]] > elementCount[uniqueElements[j]]
	})

	return uniqueElements
}

func getKeys[K comparable, V any](mymap map[K]V) []K {
	keys := make([]K, 0, len(mymap))
	for key := range mymap {
		keys = append(keys, key)
	}
	return keys
}

type Storage struct {
	Keys     map[string]int              `json:"keys"`
	RKeys    map[int]string              `json:"rkeys"`
	Index    map[string]map[int]struct{} `json:"index"`
	MaxIndex int                         `json:"maxIndex"`
}

type BaseSearchEngine struct {
	Dao Storage
}

var seg gse.Segmenter

func init() {
	seg.LoadDictEmbed("zh_s")
}

func (engine *BaseSearchEngine) Load(file string) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	bt, err := io.ReadAll(f)
	if err != nil {
		return
	}
	json.Unmarshal(bt, &engine.Dao)
}

// 是否为空
func (engine *BaseSearchEngine) IsEmpty() bool {
	if engine.Dao.Index == nil || len(engine.Dao.Index) == 0 {
		return true
	}
	return false
}

// 判断key是否存在
func (engine *BaseSearchEngine) KeyExists(key string) bool {
	if engine.Dao.Keys == nil {
		return false
	}
	_, ok := engine.Dao.Keys[key]
	return ok
}

func (engine *BaseSearchEngine) Save(file string) {
	byte, _ := json.MarshalIndent(engine.Dao, "", "    ")
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	f.Write(byte)
}

// 搜索返回keys
func (engine *BaseSearchEngine) Search(keyword string) []string {
	keyword = strings.ToLower(keyword)
	ks := seg.CutSearch(keyword, true)
	if len(ks) == 0 {
		keys := make([]string, 0, len(engine.Dao.Keys))
		for key := range engine.Dao.Keys {
			keys = append(keys, key) // 将每个 key 添加到切片中
		}
		return keys
	}
	var a = make([][]int, 0, len(ks))
	for _, k := range ks {
		k := strings.Trim(k, trim)
		if _, ok := skip[k]; ok {
			continue
		}
		if k == "" {
			continue
		}
		if b, ok := engine.Dao.Index[k]; ok {
			a = append(a, getKeys(b))
		}
	}
	c := intersection(a)
	res := make([]string, len(c))
	for i, v := range c {
		res[i] = engine.Dao.RKeys[v]
	}
	return res
}

func (engine *BaseSearchEngine) InsertOrUpdate(key, content string) {
	content = strings.ToLower(key + content)
	if engine.Dao.Keys == nil {
		engine.Dao.Keys = make(map[string]int)
	}
	if engine.Dao.RKeys == nil {
		engine.Dao.RKeys = make(map[int]string)
	}
	if engine.Dao.Index == nil {
		engine.Dao.Index = make(map[string]map[int]struct{})
	}
	segments := seg.CutSearch(content, true)

	insert := true
	// 存在
	if _, ok := engine.Dao.Keys[key]; ok {
		insert = false
	}
	if insert {
		// 新建
		engine.Dao.MaxIndex += 1
		index := engine.Dao.MaxIndex
		engine.Dao.Keys[key] = index
		engine.Dao.RKeys[index] = key
		for _, c := range segments {
			k := strings.Trim(c, trim)
			if _, ok := skip[k]; ok {
				continue
			}
			if k == "" {
				continue
			}
			if engine.Dao.Index[k] == nil {
				engine.Dao.Index[k] = make(map[int]struct{})
			}
			engine.Dao.Index[k][index] = struct{}{}
		}
	} else {
		// 修改
		// 先删除索引再新建索引
		engine.Delete(key)
		engine.InsertOrUpdate(key, content)
	}
}

func (engine *BaseSearchEngine) Delete(key string) {
	if index, ok := engine.Dao.Keys[key]; ok {
		for _, v := range engine.Dao.Index {
			delete(v, index)
		}
		delete(engine.Dao.Keys, key)
		delete(engine.Dao.RKeys, index)
	}
}
