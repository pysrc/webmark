package main

import (
	"archive/zip"
	"crypto/rand"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"webmark/utils"

	"database/sql"

	"github.com/go-ego/gse"
	_ "github.com/mattn/go-sqlite3" // 只注册驱动
	"golang.org/x/crypto/bcrypt"
)

// 数据目录
var DATA_DIR = "markdown"

// 会话持久化保存目录
var SESSIONS_DIR = "sessions"

// 会话过期时间默认一个月
var SessionExpires = 30 * 24 * time.Hour

// 数据库连接
var GDB *sql.DB

var trim = "\"\n\t `1234567890|～~!@#$%^&*()_-+=,.<>/?':;；：[]{}\\！，￥…（）—《》。？【】、”“"

var seg gse.Segmenter

func init() {
	seg.LoadDictEmbed("zh_s")
}

type UserInfo struct {
	Name     string `json:"name"`     // 用户名
	Password string `json:"password"` // 密码
}

type UserSession struct {
	SessionId string
	Name      string
	Expires   int64 // session过期时间
}

func Uuid() string {
	// 创建一个16字节的切片
	b := make([]byte, 16)

	// 从随机源中读取16字节
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	// 设置UUID版本和变体
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	// 将字节切片转换为UUID格式的字符串并打印
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// 从请求里面读json数据
func ReadJson(r *http.Request, body any) error {
	rb, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(rb, body)
	if err != nil {
		return err
	}
	return nil
}

func GetPathList(paths string, prefix string) []string {
	path := strings.TrimPrefix(paths, prefix)
	return strings.Split(path, "/")
}

type LoginRecord struct {
	LastTime int64 // 最后尝试时间
	Count    int   // 尝试次数
}

// 一天超时清理
const loginRecordTimeOut = 60 * 60 * 24

// const loginRecordTimeOut = 30

func loginErr(username string) {
	//认证失败
	var last_time, login_count int64
	err := GDB.QueryRow(`select last_time, login_count from login_record where username = ?`, username).Scan(&last_time, &login_count)
	if err != nil || login_count == 0 {
		// 插入
		_, err := GDB.Exec(`insert into login_record(username, last_time, login_count) values (?, ?, ?)`, username, time.Now().Unix(), 1)
		if err != nil {
			log.Println(err)
		}
		return
	} else {
		login_count += 1
		_, err = GDB.Exec(`update login_record set last_time = ?, login_count = ? where username = ?`, time.Now().Unix(), login_count, username)
		if err != nil {
			log.Println(err)
		}
	}
}

// 登录记录清除
func loginRecordClear() {

	rows, err := GDB.Query(`select username, last_time, login_count from login_record`)
	if err != nil {
		return
	}

	defer rows.Close()

	var deletes = make([]string, 0)

	for rows.Next() {
		var username string
		var last_time int64
		var login_count int
		err := rows.Scan(&username, &last_time, &login_count)
		if err != nil {
			continue
		}
		if login_count > 3 && time.Now().Unix()-last_time > loginRecordTimeOut {
			// 删除
			deletes = append(deletes, username)
		}
	}
	for _, username := range deletes {
		log.Println("delete", username)
		_, err = GDB.Exec(`delete from login_record where username = ?`, username)
		if err != nil {
			log.Println(err)
		}
	}
}

// 登录
type UserLogin struct {
	Username string `json:"username"` // 分组
	Password string `json:"password"` // 查询字符串
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var ul UserLogin
		if nil != ReadJson(r, &ul) {
			ErrorResponse(w, r)
			return
		}
		// 假设认证通过
		username := ul.Username
		var password string
		err := GDB.QueryRow(`select password from user_info where username = ?`, username).Scan(&password)
		if err != nil {
			// 用户不存在
			loginErr(username)
			ErrorResponse(w, r)
			return
		}
		if password == "" {
			// 用户不存在
			loginErr(username)
			ErrorResponse(w, r)
		} else {
			// 验证用户是不是在恶意尝试
			var last_time int64
			var login_count int
			err := GDB.QueryRow(`select last_time, login_count from login_record where username = ?`, username).Scan(&last_time, &login_count)
			if err == nil && login_count > 3 {
				ErrorResponseWithMsg(w, r, "fack off")
				return
			}
			if Verify(password, ul.Password) {
				// 认证通过
				var session_id = Uuid()
				// 会话过期时间
				var expires = time.Now().Add(SessionExpires)

				_, err := GDB.Exec(`insert into session_info(session_id, username, expire) values (?, ?, ?)`, session_id, username, expires.Unix())
				if err != nil {
					ErrorResponse(w, r)
					return
				}
				cookie_session_id := http.Cookie{Name: "session_id", Value: session_id, Expires: expires, Path: "/"}
				http.SetCookie(w, &cookie_session_id)
				cookie_username := http.Cookie{Name: "username", Value: username, Expires: expires, Path: "/"}
				http.SetCookie(w, &cookie_username)
				SuccessResponse(w, r, "success")
			} else {
				loginErr(username)
				ErrorResponse(w, r)
			}
		}

	} else {
		ErrorResponse(w, r)
	}
}

// 登出
func logout(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	_, err := GDB.Exec(`delete from session_info where session_id = ?`, session.SessionId)
	if err != nil {
		ErrorResponse(w, r)
		return
	}
	SuccessResponse(w, r, "logout")
}

// 用户文档组
func group_list(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}

	rows, err := GDB.Query(`select groupname from docs_group where username = ? order by create_at desc`, session.Name)
	if err != nil {
		ErrorResponse(w, r)
		return
	}
	defer rows.Close()
	var res = make([]string, 0)
	for rows.Next() {
		var groupname string
		err := rows.Scan(&groupname)
		if err != nil {
			continue
		}
		res = append(res, groupname)
	}
	SuccessResponse(w, r, res)
}

type NewGroup struct {
	Groupname string `json:"groupname"`
}

// 新建分组
// /new-group
func new_group(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	if r.Method != "POST" {
		return
	}

	var ng NewGroup
	if nil != ReadJson(r, &ng) {
		ErrorResponse(w, r)
		return
	}
	var groupname = ng.Groupname
	group_check(session.Name, groupname)
	SuccessResponse(w, r, groupname)
}

// 修改密码
// /user-password-update
type ChPass struct {
	New string `json:"new"` // 新密码
	Old string `json:"old"` // 旧密码
}

func user_password_update(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	if r.Method != "POST" {
		ErrorResponse(w, r)
		return
	}
	var cp ChPass
	if nil != ReadJson(r, &cp) {
		ErrorResponse(w, r)
		return
	}

	var oldp = cp.Old
	var newp = cp.New

	var passwd string
	err := GDB.QueryRow(`select password from user_info where username = ?`, session.Name).Scan(&passwd)
	if err != nil {
		log.Println("select user error", err)
		ErrorResponse(w, r)
		return
	}
	if Verify(passwd, oldp) {
		passwd = Genpass(newp)
		_, err := GDB.Exec(`update user_info set password = ? where username = ?`, passwd, session.Name)
		if err != nil {
			log.Println("update user error", err)
			ErrorResponse(w, r)
			return
		}
		SuccessResponse(w, r, true)
	} else {
		ErrorResponse(w, r)
	}
}

// 添加用户，仅root账户可以
// /new_user
func new_user(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	if session.Name != "root" {
		return
	}
	if r.Method != "POST" {
		return
	}
	var u UserLogin
	if nil != ReadJson(r, &u) {
		ErrorResponse(w, r)
		return
	}
	var username = u.Username
	var password = u.Password
	if nil == AddUser(username, password) {
		SuccessResponse(w, r, true)
	} else {
		ErrorResponse(w, r)
	}
}

// 新建文档
// /new-markdown/groupname/markdownname
func new_markdown(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	parts := GetPathList(r.URL.Path, "/wmapi/new-markdown/")
	var groupname = parts[0]
	var markdownname = parts[1]
	group_check(session.Name, groupname)
	var fname = DATA_DIR + "/" + session.Name + "/" + groupname + "/" + markdownname + ".md"
	_, err := os.Stat(fname)
	// 检查错误类型
	if !os.IsNotExist(err) {
		// 文件存在
		ErrorResponseWithMsg(w, r, "文件已经存在！")
		return
	}
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
		ErrorResponseWithMsg(w, r, "权限错误")
		return
	}
	fb, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorResponseWithMsg(w, r, "权限错误")
		return
	}
	file.Write(fb)
	file.Close()
	// 刷索引
	MakeIndex(session.Name, groupname, markdownname, string(fb))
	SuccessResponse(w, r, true)
}

// 修改文档
// /update-markdown/groupname/markdownname
func update_markdown(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	parts := GetPathList(r.URL.Path, "/wmapi/update-markdown/")
	var groupname = parts[0]
	var markdownname = parts[1]
	group_check(session.Name, groupname)
	var fname = DATA_DIR + "/" + session.Name + "/" + groupname + "/" + markdownname + ".md"
	_, err := os.Stat(fname)
	// 检查错误类型
	if os.IsNotExist(err) {
		// 文件不存在
		ErrorResponseWithMsg(w, r, "文件不存在！")
		return
	}
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
		ErrorResponse(w, r)
		return
	}
	fb, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorResponse(w, r)
		return
	}
	file.Write(fb)
	file.Close()
	// 刷索引
	MakeIndex(session.Name, groupname, markdownname, string(fb))
	clean_files(groupname, session.Name, markdownname)
	SuccessResponse(w, r, true)
}

// 删除文档
// /del-markdown/groupname/markdownname
func del_markdown(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	parts := GetPathList(r.URL.Path, "/wmapi/del-markdown/")
	var groupname = parts[0]
	var markdownname = parts[1]
	var fname = DATA_DIR + "/" + session.Name + "/" + groupname + "/" + markdownname
	os.RemoveAll(fname)
	os.Remove(fname + ".md")
	DeleteIndex(session.Name, groupname, markdownname)
	SuccessResponse(w, r, true)
}

// 删除分组
// /del-group/groupname
func del_group(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	parts := GetPathList(r.URL.Path, "/wmapi/del-group/")
	var groupname = parts[0]
	var fname = DATA_DIR + "/" + session.Name + "/" + groupname
	os.RemoveAll(fname)
	// 删除索引
	_, err := GDB.Exec(`delete from docs_group where username = ? and groupname = ?`, session.Name, groupname)
	if err != nil {
		log.Println("delete group error", err)
		ErrorResponse(w, r)
		return
	}
	rows, err := GDB.Query(`select doc_id from docs_info where username = ? and groupname = ?`, session.Name, groupname)
	if err != nil {
		log.Println("delete docs_info error", err)
		ErrorResponse(w, r)
		return
	}
	defer rows.Close()
	var doc_ids = make([]int, 0)
	for rows.Next() {
		var doc_id int
		err := rows.Scan(&doc_id)
		if err != nil {
			continue
		}
		doc_ids = append(doc_ids, doc_id)
	}
	for _, doc_id := range doc_ids {
		_, err := GDB.Exec(`delete from docs where rowid = ?`, doc_id)
		if err != nil {
			log.Println("delete docs error", err)
			ErrorResponse(w, r)
			return
		}
	}
	_, err = GDB.Exec(`delete from docs_info where username = ? and groupname = ?`, session.Name, groupname)
	if err != nil {
		log.Println("delete docs_info error", err)
		ErrorResponse(w, r)
		return
	}
	SuccessResponse(w, r, true)
}

// 用户检测，不存在就创建目录
func user_check(name string) {
	var user_dir = DATA_DIR + "/" + name
	if _, err := os.Stat(user_dir); os.IsNotExist(err) {
		err := os.MkdirAll(user_dir, 0755)
		if err != nil {
			return
		}
	}
}

// 分组检测
func group_check(username string, group string) {
	user_check(username)
	var user_dir = DATA_DIR + "/" + username + "/" + group
	if _, err := os.Stat(user_dir); os.IsNotExist(err) {
		err := os.MkdirAll(user_dir, 0755)
		if err != nil {
			return
		}
	}
	// 创建数据库记录
	var count int
	err := GDB.QueryRow(`select count(1) from docs_group where username = ? and groupname = ?`, username, group).Scan(&count)
	if err != nil {
		log.Println("group_check error", err)
		return
	}
	if count == 0 {
		_, err := GDB.Exec(`insert into docs_group (username, groupname, create_at) values (?, ?, ?)`, username, group, time.Now().Unix())
		if err != nil {
			log.Println("group_check error", err)
			return
		}
	}
}

// 压缩文件夹
func ZipDir(src_dir string, dst_writer io.Writer) {
	// 打开：zip文件
	archive := zip.NewWriter(dst_writer)
	defer archive.Close()

	// 遍历路径信息
	filepath.Walk(src_dir, func(path string, info os.FileInfo, _ error) error {

		path = strings.ReplaceAll(path, `\`, `/`)
		// 如果是源路径，提前进行下一个遍历
		if path == src_dir {
			return nil
		}

		// 获取：文件头信息
		header, _ := zip.FileInfoHeader(info)

		header.Name = strings.TrimPrefix(path, src_dir+"/")

		// 判断：文件是不是文件夹
		if info.IsDir() {
			header.Name += `/`
		} else {
			// 设置：zip的文件压缩算法
			header.Method = zip.Deflate
		}

		// 创建：压缩包头部信息
		writer, _ := archive.CreateHeader(header)
		if !info.IsDir() {
			file, _ := os.Open(path)
			defer file.Close()
			io.Copy(writer, file)
		}
		return nil
	})
}

// 文档导出
// /export
// /export/groupname
// /export/groupname/markdownname
func export(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	user_check(session.Name)
	var username = session.Name
	if r.URL.Path == "/wmapi/export/" {
		// 导出整个用户的文档
		var fname = username + ".zip"
		w.Header().Add("Content-Disposition", "attachment; filename="+fname)
		w.Header().Add("Content-Type", "application/octet-stream")
		ZipDir(DATA_DIR+"/"+username, w)
	} else {
		var restname = strings.TrimPrefix(r.URL.Path, "/wmapi/export/")
		var rts = strings.Split(restname, "/")
		if len(rts) == 1 {
			// 导出某个组的文档
			var fname = rts[0] + ".zip"
			w.Header().Add("Content-Disposition", "attachment; filename="+fname)
			w.Header().Add("Content-Type", "application/octet-stream")
			ZipDir(DATA_DIR+"/"+username+"/"+rts[0], w)
		} else if len(rts) == 2 {
			var fname = rts[1] + ".zip"
			var mdinfo = DATA_DIR + "/" + username + "/" + rts[0] + "/" + rts[1]
			w.Header().Add("Content-Disposition", "attachment; filename="+fname)
			w.Header().Add("Content-Type", "application/octet-stream")
			archive := zip.NewWriter(w)
			defer archive.Close()
			finfo, err := os.Stat(mdinfo + ".md")
			if os.IsNotExist(err) {
				return
			}
			header, _ := zip.FileInfoHeader(finfo)
			header.Name = rts[1] + ".md"
			header.Method = zip.Deflate
			writer, _ := archive.CreateHeader(header)
			file, _ := os.Open(mdinfo + ".md")
			defer file.Close()
			io.Copy(writer, file)
			// 判断是否有附件信息
			dirinfo, err := os.Stat(mdinfo)
			if os.IsNotExist(err) {
				return
			}
			header, _ = zip.FileInfoHeader(dirinfo)
			header.Name = rts[1]
			archive.CreateHeader(header)
			// 遍历文件夹
			f, err := os.Open(mdinfo)
			if err != nil {
				fmt.Println("打开目录时出错：", err)
				return
			}
			defer f.Close()
			files, err := f.Readdir(-1)
			if err != nil {
				fmt.Println("读取目录时出错：", err)
				return
			}
			for _, fi := range files {
				header, _ := zip.FileInfoHeader(fi)
				header.Name = rts[1] + "/" + fi.Name()
				header.Method = zip.Deflate
				writer, _ := archive.CreateHeader(header)
				file, _ := os.Open(mdinfo + "/" + fi.Name())
				io.Copy(writer, file)
				file.Close()
			}
		}
	}
}

/*
文件上传
*/
// /upload/groupname/markdownname
func upload(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	user_check(session.Name)
	parts := GetPathList(r.URL.Path, "/wmapi/upload/")
	var groupname = parts[0]
	var markdownname = parts[1]

	var work_dir = DATA_DIR + "/" + session.Name + "/" + groupname + "/" + markdownname
	// 文件夹不存在就创建
	if _, err := os.Stat(work_dir); os.IsNotExist(err) {
		err := os.MkdirAll(work_dir, 0755)
		if err != nil {
			ErrorResponse(w, r)
			return
		}
	}

	// 从请求中获取文件
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		ErrorResponse(w, r)
		return
	}
	defer file.Close()

	// 创建一个新文件
	newFile, err := os.Create(work_dir + "/" + header.Filename)
	if err != nil {
		log.Println(err)
		ErrorResponse(w, r)
		return
	}
	defer newFile.Close()

	// 将上传的文件内容复制到新文件中
	_, err = io.Copy(newFile, file)
	if err != nil {
		log.Println(err)
		ErrorResponse(w, r)
		return
	}

	// 返回上传成功的信息
	SuccessResponse(w, r, true)
}

/*
清理多余文件
*/
func clean_files(groupname, username, markdown string) bool {
	var _path = DATA_DIR + "/" + username + "/" + groupname + "/" + markdown
	// 读文件
	var _md_path = _path + ".md"
	// 提取出所有文件相关地址
	_bytes, err := os.ReadFile(_md_path)
	if err != nil {
		return false
	}
	var _content = string(_bytes)
	// 读出文件出中所有的文件
	if _sub_fs, err := os.ReadDir(_path); os.IsNotExist(err) {
		return true
	} else if err != nil {
		return true
	} else {
		// 删除不在文章中存在的文件
		for _, _fs := range _sub_fs {
			if !_fs.IsDir() {
				if !strings.Contains(_content, markdown+"/"+_fs.Name()) {
					os.Remove(_path + "/" + _fs.Name())
				}
			}
		}
	}
	return true
}

type ResponseBase struct {
	Ok   bool   `json:"ok"`   // 是否成功
	Data any    `json:"data"` // 数据
	Msg  string `json:"msg"`  // 信息
}

func ErrorResponse(w http.ResponseWriter, r *http.Request) {
	var res = ResponseBase{
		Ok:  false,
		Msg: "",
	}
	j, _ := json.Marshal(res)
	w.Header().Add("content-type", "application/json")
	w.Write(j)
}

func ErrorResponseWithMsg(w http.ResponseWriter, r *http.Request, msg string) {
	var res = ResponseBase{
		Ok:  false,
		Msg: msg,
	}
	j, _ := json.Marshal(res)
	w.Header().Add("content-type", "application/json")
	w.Write(j)
}

func SuccessResponse(w http.ResponseWriter, r *http.Request, data any) {
	var res = ResponseBase{
		Ok:   true,
		Data: data,
	}
	j, _ := json.Marshal(res)
	w.Header().Add("content-type", "application/json")
	w.Write(j)
}

func Groupname(r *http.Request) string {
	var cookie, err = r.Cookie("groupname")
	if err != nil {
		return ""
	}
	s, _ := url.QueryUnescape(cookie.Value)
	return s
}

func Auth(w http.ResponseWriter, r *http.Request) (bool, *UserSession) {
	// 需要权限控制的
	var cookie, err = r.Cookie("session_id")
	if err != nil {
		ErrorResponse(w, r)
		return false, nil
	} else {
		var session_id = cookie.Value
		var username string
		var expire int64
		err = GDB.QueryRow(`select username, expire from session_info where session_id = ?`, session_id).Scan(&username, &expire)
		if err != nil {
			log.Println(err)
			ErrorResponse(w, r)
			return false, nil
		}
		// 校验session是否过期
		if expire < time.Now().Unix() {
			_, err = GDB.Exec(`delete from session_info where session_id = ?`, session_id)
			if err != nil {
				ErrorResponse(w, r)
				return false, nil
			}
			return false, nil
		}
		if username == "" {
			// 用户未登录
			ErrorResponse(w, r)
			return false, nil
		} else {
			// 用户已登录
			return true, &UserSession{
				Expires:   expire,
				Name:      username,
				SessionId: session_id,
			}
		}

	}

}

// 搜索
type SearchInput struct {
	Group string `json:"group"` // 分组
	Query string `json:"query"` // 查询字符串
}

func search_detail(w http.ResponseWriter, r *http.Request) {
	var input SearchInput
	if nil != ReadJson(r, &input) {
		ErrorResponse(w, r)
		return
	}
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	sps := splitWord(input.Query)
	sqls := `select di.title from docs d, docs_info di where di.doc_id = d.rowid and di.username = ? and di.groupname = ?`
	q := strings.Join(sps, " AND ")
	if len(sps) > 0 {
		sqls += " AND d.docs MATCH ? order by di.create_at DESC"
		rows, err := GDB.Query(sqls, session.Name, input.Group, q)
		if err != nil {
			ErrorResponse(w, r)
			return
		}
		defer rows.Close()
		var res = make([]string, 0)
		for rows.Next() {
			var title string
			err := rows.Scan(&title)
			if err != nil {
				ErrorResponse(w, r)
				return
			}
			res = append(res, title)
		}
		SuccessResponse(w, r, res)
	} else {
		sqls += " order by di.create_at DESC"
		rows, err := GDB.Query(sqls, session.Name, input.Group)
		if err != nil {
			ErrorResponse(w, r)
			return
		}
		defer rows.Close()
		var res = make([]string, 0)
		for rows.Next() {
			var title string
			err := rows.Scan(&title)
			if err != nil {
				ErrorResponse(w, r)
				return
			}
			res = append(res, title)
		}
		SuccessResponse(w, r, res)
	}
}

// 分词
func splitWord(content string) []string {
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
	sps := make([]string, 0, len(tm))
	for k := range tm {
		sps = append(sps, k)
	}
	return sps
}

// 建索引&刷新索引
func MakeIndex(user, group, title, content string) {
	// 判断文本是否存在
	var doc_id int
	err := GDB.QueryRow(`select doc_id from docs_info where groupname = ? and username = ? and title = ?`, group, user, title).Scan(&doc_id)
	if err != nil && doc_id == 0 {
		// 插入文档描述
		_, err = GDB.Exec(`insert into docs_info(groupname, title, username, create_at) values (?, ?, ?, ?)`, group, title, user, time.Now().Unix())
		if err != nil {
			log.Println("MakeIndex insert error", err)
			return
		}
		err = GDB.QueryRow(`select doc_id from docs_info where groupname = ? and username = ? and title = ?`, group, user, title).Scan(&doc_id)
		if err != nil {
			log.Println("MakeIndex query error", err)
			return
		}
		// 新增索引
		// 分词
		tt := splitWord(title)
		tc := splitWord(content)
		_, err = GDB.Exec(`insert into docs(rowid, title, content) values (?, ?, ?)`, doc_id, strings.Join(tt, " "), strings.Join(tc, " "))
		if err != nil {
			log.Println("MakeIndex insert error", err)
			return
		}
	} else {
		// 修改时间
		_, err = GDB.Exec(`update docs_info set create_at = ? where doc_id = ?`, time.Now().Unix(), doc_id)
		if err != nil {
			log.Println("MakeIndex update error", err)
			return
		}
		// 删除索引
		_, err = GDB.Exec(`delete from docs where rowid = ?`, doc_id)
		if err != nil {
			log.Println("MakeIndex delete error", err)
			return
		}
		// 新增索引
		tt := splitWord(title)
		tc := splitWord(content)
		_, err = GDB.Exec(`insert into docs(rowid, title, content) values (?, ?, ?)`, doc_id, strings.Join(tt, " "), strings.Join(tc, " "))
		if err != nil {
			log.Println("MakeIndex insert error", err)
			return
		}
	}
}

// 删除索引
func DeleteIndex(user, group, title string) {
	var doc_id int
	err := GDB.QueryRow(`select doc_id from docs_info where groupname = ? and username = ? and title = ?`, group, user, title).Scan(&doc_id)
	if err != nil {
		log.Println("DeleteIndex query error", err)
		return
	}
	if doc_id != 0 {
		// 删除索引
		_, err = GDB.Exec(`delete from docs where rowid = ?`, doc_id)
		if err != nil {
			log.Println("DeleteIndex delete error", err)
			return
		}
		// 删除文档描述
		_, err = GDB.Exec(`delete from docs_info where doc_id = ?`, doc_id)
		if err != nil {
			log.Println("DeleteIndex delete error", err)
			return
		}
	}
}

func AddUser(name string, password string) error {
	// 添加用户
	// 判断用户是否存在
	var count int
	err := GDB.QueryRow(`select count(1) from user_info where username = ?`, name).Scan(&count)
	if err != nil {
		log.Println("AddUser query error", err)
		return err
	}
	if count > 0 {
		return errors.New("user exists")
	}

	ep := Genpass(password)
	// 插入数据库
	_, err = GDB.Exec(`insert into user_info (username, password) values (?, ?)`, name, ep)
	if err != nil {
		return err
	}
	return nil
}

func init_work() {
	err := AddUser("root", "root")
	if err == nil {
		log.Println("初始化用户名: root")
		log.Println("初始化密码: root")
	}
}

// 自动扫描已有文件夹创建索引
func createIndex(username string) {
	log.Println("开始创建索引...")
	defer log.Println("创建索引完成")
	// 扫描用户文件夹
	log.Println("开始处理用户: ", username)
	// 扫描用户组
	groups, err := utils.ListImmediateSubDirectories(filepath.Join(DATA_DIR, username))
	if err != nil {
		log.Println("createIndex error: ", err)
		return
	}
	for _, group := range groups {
		log.Println("开始处理用户组: ", group.Name())
		// 判断组是否存在
		group_check(username, group.Name())
		// 列出所有md文件
		files, err := utils.ListImmediateSubFiles(filepath.Join(DATA_DIR, username, group.Name()))
		if err != nil {
			log.Println("列出所有md文件失败: ", err)
			continue
		}
		for _, file := range files {
			log.Println("开始处理文件: ", file.Name())
			if !strings.HasSuffix(file.Name(), ".md") {
				continue
			}
			title := strings.TrimSuffix(file.Name(), ".md")
			var count int
			err := GDB.QueryRow(`select count(1) from docs_info where groupname = ? and username = ? and title = ?`, group.Name(), username, title).Scan(&count)
			if err != nil {
				log.Println("createIndex error: ", err)
				continue
			}
			if count == 0 {
				// 创建索引
				content, err := os.ReadFile(filepath.Join(DATA_DIR, username, group.Name(), file.Name()))
				if err != nil {
					log.Println("createIndex error: ", err)
					continue
				}
				log.Println("开始创建索引: ", title)
				MakeIndex(username, group.Name(), title, string(content))
			}
		}
	}

}

func SessionClear() {
	// 把超时的session踢出去
	rows, err := GDB.Query(`select session_id, username, expire from session_info`)
	if err != nil {
		log.Println("session job error: ", err)
		return
	}
	defer rows.Close()

	var sessionids = make([]string, 0)
	for rows.Next() {
		var session_id, username string
		var expire int64
		err := rows.Scan(&session_id, &username, &expire)
		if err != nil {
			log.Println("session loop error: ", err)
			continue
		}
		if expire < time.Now().Unix() {
			// session 超时
			sessionids = append(sessionids, session_id)
		}
	}
	for _, session_id := range sessionids {
		_, err = GDB.Exec(`delete from session_info where session_id = ?`, session_id)
		if err != nil {
			log.Println("session delete error: ", err)
			continue
		}
	}
}

// 定时任务
func Job() {
	var dur = 1 * time.Hour
	// var dur = 1 * time.Second
	t := time.NewTimer(dur)
	for {
		<-t.C
		t.Reset(dur)
		SessionClear()
		// 检查恶意登录
		loginRecordClear()
	}
}

func Genpass(passwd string) string {
	// 生成哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hashedPassword)
}

func Verify(hashedPassword string, enteredPassword string) bool {
	// 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(enteredPassword))
	if err == nil {
		return true
	} else {
		return false
	}
}

//go:embed page/*
var staticFiles embed.FS

// 判断是否静态文件
func IsStatic(name string) bool {
	if name == "/" ||
		strings.Contains(name, "/static/") ||
		strings.HasSuffix(name, "index.html") ||
		strings.HasSuffix(name, "asset-manifest.json") ||
		strings.HasSuffix(name, "favicon.ico") ||
		strings.HasSuffix(name, "logo192.png") ||
		strings.HasSuffix(name, "logo512.png") ||
		strings.HasSuffix(name, "manifest.json") ||
		strings.HasSuffix(name, "robots.txt") {
		return true
	}
	return false
}

type StaticEntry struct {
	StaticFs   http.FileSystem
	MarkdownFs http.FileSystem
}

func (fs StaticEntry) Open(name string) (http.File, error) {
	if IsStatic(name) {
		file, err := fs.StaticFs.Open(name)
		if err == nil {
			return file, nil
		}
	} else {
		file, err := fs.MarkdownFs.Open(name)
		if err == nil {
			return file, nil
		}
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

func auth_static(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsStatic(r.URL.Path) {
			http.StripPrefix("/", next).ServeHTTP(w, r)
		} else {
			// 登录的
			log.Println(r.URL.Path)
			if suc, se := Auth(w, r); suc {
				log.Println(r.URL.Path)
				s := strings.TrimPrefix(r.URL.Path, "/wmapi/markdown")
				p := "/" + se.Name + "/" + s
				r.URL.Path = p
				log.Println(r.URL.Path)
				next.ServeHTTP(w, r)
				return
			}
		}
	})
}

func createTable() error {

	_, err := GDB.Exec(`CREATE TABLE IF NOT EXISTS user_info (
		username  varchar(100),
		password varchar(100)
		)`)
	if err != nil {
		log.Println("createTable error", err)
		return err
	}

	_, err = GDB.Exec(`CREATE TABLE IF NOT EXISTS session_info (session_id varchar (100), username varchar (100), expire INTEGER)`)
	if err != nil {
		log.Println("createTable error", err)
		return err
	}

	_, err = GDB.Exec(`CREATE TABLE IF NOT EXISTS login_record (username varchar (100), last_time INTEGER, login_count INTEGER)`)
	if err != nil {
		log.Println("createTable error", err)
		return err
	}

	_, err = GDB.Exec(`CREATE TABLE IF NOT EXISTS docs_group(groupname varchar(100), username varchar(100), create_at INTEGER)`)
	if err != nil {
		log.Println("createTable error", err)
		return err
	}

	_, err = GDB.Exec(`CREATE TABLE IF NOT EXISTS docs_info(doc_id INTEGER PRIMARY KEY AUTOINCREMENT, groupname varchar(100), title varchar(100), username  varchar(100), create_at INTEGER)`)

	if err != nil {
		log.Println("createTable error", err)
		return err
	}

	_, err = GDB.Exec(`CREATE INDEX IF NOT EXISTS docs_username ON docs_info(username)`)

	if err != nil {
		log.Println("createTable error", err)
		return err
	}

	_, err = GDB.Exec(`CREATE INDEX IF NOT EXISTS docs_groupname ON docs_info(groupname)`)

	if err != nil {
		log.Println("createTable error", err)
		return err
	}

	_, err = GDB.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS docs USING fts5(title, content)`)

	if err != nil {
		log.Println("createTable error", err)
		return err
	}

	return nil
}

func updateIndex(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	go createIndex(session.Name)
	SuccessResponse(w, r, true)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var bind, passwd string
	var genpass bool
	flag.BoolVar(&genpass, "genpass", false, "生成hash密码")
	flag.StringVar(&passwd, "gen-password", "markdown", "需要加密的密码")
	flag.StringVar(&DATA_DIR, "data", "markdown", "文档存储目录")
	flag.StringVar(&bind, "bind", "127.0.0.1:11990", "绑定host与端口信息")
	flag.StringVar(&SESSIONS_DIR, "sessions", "sessions", "会话持久化目录")
	flag.Parse()

	if genpass {
		p := Genpass(passwd)
		fmt.Println(p)
		return
	}

	db, err := sql.Open("sqlite3", "./webmark.db?_journal_mode=WAL") // 文件不存在会自动创建
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	GDB = db

	createTable()

	log.Println("浏览器地址：http://" + bind)
	init_work()
	go Job()
	webroot, _ := fs.Sub(staticFiles, "page")
	multiDir := StaticEntry{StaticFs: http.FS(webroot), MarkdownFs: http.Dir(DATA_DIR + "/")}
	http.Handle("/", auth_static(http.FileServer(multiDir)))
	// http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("page/"))))
	// http.Handle("/markdown/", auth_markdown(http.FileServer(http.Dir(DATA_DIR+"/"))))
	http.HandleFunc("/wmapi/upload/", upload)
	http.HandleFunc("/wmapi/login", login)
	http.HandleFunc("/wmapi/logout", logout)
	http.HandleFunc("/wmapi/group-list", group_list)
	http.HandleFunc("/wmapi/new-group", new_group)
	http.HandleFunc("/wmapi/new-markdown/", new_markdown)
	http.HandleFunc("/wmapi/update-markdown/", update_markdown)
	http.HandleFunc("/wmapi/del-markdown/", del_markdown)
	http.HandleFunc("/wmapi/del-group/", del_group)
	http.HandleFunc("/wmapi/user-password-update", user_password_update)
	http.HandleFunc("/wmapi/new-user", new_user)
	http.HandleFunc("/wmapi/export/", export)
	http.HandleFunc("/wmapi/search-detail", search_detail)
	// 刷新索引
	http.HandleFunc("/wmapi/update-index", updateIndex)
	server := http.Server{Addr: bind}
	server.ListenAndServe()
}
