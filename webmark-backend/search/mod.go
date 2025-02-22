package search

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/go-ego/gse"
)

type MetaInfo struct {
	IdTitle map[int64]string `json:"id-title"`
	TitleId map[string]int64 `json:"title-id"`
	Maxid   int64            `json:"maxid"`
}

var IndexDir = ".search"

var trim = "\"\n\t `1234567890|～~!@#$%^&*()_-+=,.<>/?':;；：[]{}\\！，￥…（）—《》。？【】、”“"

var seg gse.Segmenter

func init() {
	seg.LoadDictEmbed("zh_s")
}

// 是否空目录
func isDirEmpty(dirPath string) (bool, error) {
	// 打开目录
	dir, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}
	defer dir.Close()

	// 读取目录下的所有文件和子目录
	fileInfos, err := dir.Readdir(1)
	if err != nil {
		// 空文件夹
		if err == io.EOF {
			return true, nil
		}
		return false, err
	}

	// 如果目录为空，返回true；否则返回false
	return len(fileInfos) == 0, nil
}

// 目录为空就生成
func handEmptyDir(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}

// 写文件
func writeFile(file_path string, content []byte) {
	file, err := os.OpenFile(file_path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	file.Write(content)
}

var hasher = sha1.New()

func getSha1(data []byte) []byte {
	// 创建一个新的SHA-1哈希对象
	hasher.Reset()
	// 写入数据到哈希对象
	hasher.Write(data)
	return hasher.Sum(nil)
}

func readFile(fpath string) ([]byte, error) {
	rfile, err := os.OpenFile(fpath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer rfile.Close()
	return io.ReadAll(rfile)
}

type LowSearch struct {
	meta     *MetaInfo
	BasePath string // 文档地址
}

// 写索引
func (search *LowSearch) writeIndex(hash []byte, data []byte) {
	dirs := fmt.Sprintf("%s/%02x", search.getIndexPath(), hash[0])
	handEmptyDir(dirs)
	datapath := hex.EncodeToString(hash[1:])
	writeFile(fmt.Sprintf("%s/%s", dirs, datapath), data)
}

// 读索引
func (search *LowSearch) readIndex(hash []byte) ([]byte, error) {
	dirs := fmt.Sprintf("%s/%02x", search.getIndexPath(), hash[0])
	datapath := hex.EncodeToString(hash[1:])
	index_file := fmt.Sprintf("%s/%s", dirs, datapath)
	return readFile(index_file)
}

// 删除索引
func (search *LowSearch) deleteIndex(hash []byte, id int64) error {
	dirs := fmt.Sprintf("%s/%02x", search.getIndexPath(), hash[0])
	datapath := hex.EncodeToString(hash[1:])
	index_file := fmt.Sprintf("%s/%s", dirs, datapath)
	idata, err := readFile(index_file)
	if err != nil {
		return err
	}
	// 找到id删除
	idb := []byte(fmt.Sprint(id))
	var tempbuffer bytes.Buffer
	var restbuffer bytes.Buffer
	for _, v := range idata {
		if v == 0 {
			temp := tempbuffer.Bytes()
			tempbuffer.Reset()
			if bytes.Equal(temp, idb) {
				continue
			} else {
				restbuffer.Write(temp)
				restbuffer.WriteByte(0)
			}
		} else {
			tempbuffer.WriteByte(v)
		}
	}
	var d = restbuffer.Bytes()
	if len(d) == 0 {
		os.Remove(index_file)
	} else {
		search.writeIndex(hash, d)
	}
	if ise, _ := isDirEmpty(dirs); ise {
		os.Remove(dirs)
	}
	return nil
}

// 获取搜索根目录
func (search *LowSearch) getRootPath() string {
	return fmt.Sprintf("%s/%s", search.BasePath, IndexDir)
}

func (search *LowSearch) getIndexPath() string {
	return fmt.Sprintf("%s/%s", search.getRootPath(), "index")
}

func (search *LowSearch) getMetaPath() string {
	return fmt.Sprintf("%s/%s", search.getRootPath(), "meta.json")
}

func (search *LowSearch) getMetainfo() (*MetaInfo, error) {
	if search.meta != nil {
		return search.meta, nil
	}
	data, err := readFile(search.getMetaPath())
	if err != nil {
		if os.IsNotExist(err) {
			// 不存在，新建
			var meta = MetaInfo{
				IdTitle: make(map[int64]string),
				TitleId: make(map[string]int64),
				Maxid:   0,
			}
			data, _ = json.Marshal(meta)
			writeFile(search.getMetaPath(), data)
			search.meta = &meta
			return search.meta, nil
		} else {
			return nil, err
		}
	} else {
		var meta MetaInfo
		json.Unmarshal(data, &meta)
		search.meta = &meta
		return search.meta, nil
	}
}

func (search *LowSearch) getKeywordPath() string {
	return fmt.Sprintf("%s/%s", search.getRootPath(), "keyword")
}

func (search *LowSearch) getNextId() int64 {
	meta, _ := search.getMetainfo()
	meta.Maxid += 1
	return meta.Maxid
}

func (search *LowSearch) saveMeta() {
	data, _ := json.Marshal(search.meta)
	writeFile(search.getMetaPath(), data)
}

// 初始化
func (search *LowSearch) Init() {
	handEmptyDir(search.getRootPath())
	handEmptyDir(search.getKeywordPath())
	handEmptyDir(search.getIndexPath())
}

// 分词
func splitWord(content string) map[string]struct{} {
	// 全转小写
	content = strings.ToLower(content)
	// 分词
	segments := seg.CutSearch(content, true)
	// 去重
	var tm = make(map[string]struct{}, len(segments))
	for _, s := range segments {
		k := strings.Trim(s, trim)
		if k == "" {
			continue
		}
		if len(k) > 10 {
			continue
		}
		tm[s] = struct{}{}
	}
	return tm
}

// 插入或更新文档
func (search *LowSearch) InsertOrUpdate(title, content string) {
	content = strings.ToLower(title + " " + content)
	// 查看文档是否存在
	meta, _ := search.getMetainfo()
	id := meta.TitleId[title]
	if id == 0 {
		// 不存在，插入
		// 生成ID
		id := search.getNextId()
		// 关联title
		meta.IdTitle[id] = title
		meta.TitleId[title] = id
		// 分词
		var sw = splitWord(title + " " + content)
		// 写入keyword文件
		// 写索引
		var buffer bytes.Buffer
		for key := range sw {
			buffer.WriteString(key)
			buffer.WriteByte(0)
			hash := getSha1([]byte(key))
			// 读索引
			idata, err := search.readIndex(hash)
			dc := append([]byte(fmt.Sprint(id)), 0)
			if os.IsNotExist(err) {
				// 新建索引
				search.writeIndex(hash, dc)
			} else {
				// 修改索引
				idata = append(idata, dc...)
				search.writeIndex(hash, idata)
			}
		}
		// 写入id与索引关联文件
		kp := fmt.Sprintf("%s/%d", search.getKeywordPath(), id)
		writeFile(kp, buffer.Bytes())
		// 写meta
		search.saveMeta()
	} else {
		// 存在，修改
		// 先删除
		search.Delete(title)
		// 再插入
		search.InsertOrUpdate(title, content)
	}

}

// 删除文档
func (search *LowSearch) Delete(title string) {
	meta, _ := search.getMetainfo()
	id := meta.TitleId[title]
	delete(meta.IdTitle, id)
	delete(meta.TitleId, title)
	// 删索引文件
	// 获取全部索引
	kwp := fmt.Sprintf("%s/%d", search.getKeywordPath(), id)
	idata, err := readFile(kwp)
	os.Remove(kwp)
	if err != nil {
		return
	}
	bb := bytes.Split(idata, []byte{0})
	// 遍历文件和子目录，筛选出子目录
	for _, key := range bb {
		hash := getSha1(key)
		search.deleteIndex(hash, id)
	}
	search.saveMeta()
}

// 计算整数切片的交集，并返回按照次数递减排序的列表（不包含重复元素）
func intersection(slices [][]int64) []int64 {
	if len(slices) == 0 {
		return nil
	}

	// 创建一个映射来跟踪元素的出现次数
	elementCount := make(map[int64]int)

	// 遍历第一个切片，记录元素出现次数
	for _, element := range slices[0] {
		elementCount[element]++
	}

	// 遍历剩余切片，更新元素出现次数
	for _, slice := range slices[1:] {
		currentElementCount := make(map[int64]int)

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
	var uniqueElements []int64
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

// 查询文档，返回标题列表
func (search *LowSearch) Search(keyword string) []string {
	// 分词
	var sw = splitWord(keyword)
	if len(sw) == 0 {
		meta, _ := search.getMetainfo()
		rkeys := make([]int64, 0, len(meta.IdTitle))
		for key := range meta.IdTitle {
			rkeys = append(rkeys, key) // 将每个 key 添加到切片中
		}
		// 从大到小排序
		sort.Slice(rkeys, func(i, j int) bool {
			return rkeys[i] > rkeys[j]
		})

		keys := make([]string, 0, len(meta.IdTitle))
		for _, key := range rkeys {
			keys = append(keys, meta.IdTitle[key]) // 将每个 key 添加到切片中
		}
		return keys
	} else {
		var a = make([][]int64, 0, len(sw))
		for k := range sw {
			hash := getSha1([]byte(k))
			idata, err := search.readIndex(hash)
			if err != nil {
				continue
			}
			ta := make([]int64, 0, 10)
			var t int64 = 0
			for _, v := range idata {
				if v == 0 {
					ta = append(ta, t)
					t = 0
				} else {
					t = t*10 + int64(v-'0')
				}
			}
			a = append(a, ta)
		}
		c := intersection(a)
		res := make([]string, len(c))
		// 从大到小排序
		sort.Slice(c, func(i, j int) bool {
			return c[i] > c[j]
		})
		meta, _ := search.getMetainfo()
		for i, v := range c {
			res[i] = meta.IdTitle[v]
		}
		return res
	}
}
