package main

import (
	"archive/zip"
	"crypto/rand"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 数据目录
var DATA_DIR = "markdown"

// 用户信息目录
var USERS_DIR = "users"

// 会话持久化保存目录
var SESSIONS_DIR = "sessions"

// 会话过期时间默认一个月
var SessionExpires = 30 * 24 * time.Hour

type UserInfo struct {
	Name      string                `json:"name"`       // 用户名
	Password  string                `json:"password"`   // 密码
	Group     map[string]*GroupInfo `json:"group"`      // 文档组
	GroupSort []string              `json:"group_sort"` // 按时间先后顺序排列的组名
}

type GroupInfo struct {
	Name      string   `json:"name"`
	Markdowns []string `json:"markdowns"`
}

type UserSession struct {
	SessionId string
	Name      string
	Expires   int64 // session过期时间
}

var user_map = make(map[string]*UserInfo)

var session_map = make(map[string]*UserSession)

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

func GetPathList(paths string, prefix string) []string {
	path := strings.TrimPrefix(paths, prefix)
	return strings.Split(path, "/")
}

// 登录
func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 假设认证通过
		username := r.PostFormValue("username")
		var usr = user_map[username]
		if usr == nil {
			// 用户不存在
			AuthError(w, r)
		} else {
			if usr.Password == r.PostFormValue("password") {
				// 认证通过
				var session_id = Uuid()
				// 会话过期时间
				var expires = time.Now().Add(SessionExpires)
				session_map[session_id] = &UserSession{
					SessionId: session_id,
					Name:      username,
					Expires:   expires.Unix(),
				}
				session_save(session_id)
				cookie_session_id := http.Cookie{Name: "session_id", Value: session_id, Expires: expires}
				http.SetCookie(w, &cookie_session_id)
				cookie_username := http.Cookie{Name: "username", Value: username, Expires: expires}
				http.SetCookie(w, &cookie_username)
				http.Redirect(w, r, "user_main.html", http.StatusSeeOther)
			} else {
				//认证失败
				AuthError(w, r)
			}
		}

	} else {
		AuthError(w, r)
	}
}

// 登出
func logout(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	delete(session_map, session.SessionId)
	os.Remove(SESSIONS_DIR + "/" + session.SessionId + ".json")
	AuthError(w, r)
}

// 用户文档组
func group_list(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	var user = user_map[session.Name]
	w.Header().Add("content-type", "application/json")
	ret := make(map[string][]string)
	ret["group"] = user.GroupSort
	res, _ := json.Marshal(ret)
	w.Write(res)
}

// 文档列表
func markdown_list(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	var user = user_map[session.Name]
	var group = r.FormValue("group")
	w.Header().Add("content-type", "application/json")
	var mds = user.Group[group]
	res, _ := json.Marshal(mds)
	w.Write(res)
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
	var groupname = r.PostFormValue("groupname")
	group_check(session.Name, groupname)
	http.Redirect(w, r, "group_main.html?groupname="+groupname, http.StatusSeeOther)
}

// 修改密码
// /user-password-update
func user_password_update(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	if r.Method != "POST" {
		return
	}
	var oldp = r.PostFormValue("old")
	var newp = r.PostFormValue("new")
	if user_map[session.Name].Password == oldp {
		user_map[session.Name].Password = newp
		cache_save(session.Name)
		w.Write([]byte("YES"))
	} else {
		w.Write([]byte("NO"))
	}
}

// 新建、修改文档
// /new-markdown/groupname/markdownname
func new_markdown(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	parts := GetPathList(r.URL.Path, "/new-markdown/")
	var groupname = parts[0]
	var markdownname = parts[1]
	group_check(session.Name, groupname)
	var fname = DATA_DIR + "/" + session.Name + "/" + groupname + "/" + markdownname + ".md"
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	io.Copy(file, r.Body)
	// 刷盘
	var mds = user_map[session.Name].Group[groupname].Markdowns
	var inside = false
	for _, k := range mds {
		if k == markdownname {
			inside = true
			break
		}
	}
	if !inside {
		user_map[session.Name].Group[groupname].Markdowns = append(user_map[session.Name].Group[groupname].Markdowns, markdownname)
		cache_save(session.Name)
	}
}

// 删除文档
// /del-markdown/groupname/markdownname
func del_markdown(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	parts := GetPathList(r.URL.Path, "/del-markdown/")
	var groupname = parts[0]
	var markdownname = parts[1]
	var fname = DATA_DIR + "/" + session.Name + "/" + groupname + "/" + markdownname
	var mds = user_map[session.Name].Group[groupname].Markdowns
	for i := 0; i < len(mds); i++ {
		if mds[i] == markdownname {
			os.RemoveAll(fname)
			os.Remove(fname + ".md")
			mds = append(mds[:i], mds[i+1:]...)
			user_map[session.Name].Group[groupname].Markdowns = mds
			cache_save(session.Name)
			return
		}
	}
}

// 删除分组
// /del-group/groupname
func del_group(w http.ResponseWriter, r *http.Request) {
	var suc, session = Auth(w, r)
	if !suc {
		return
	}
	parts := GetPathList(r.URL.Path, "/del-group/")
	var groupname = parts[0]
	var fname = DATA_DIR + "/" + session.Name + "/" + groupname
	os.RemoveAll(fname)
	delete(user_map[session.Name].Group, groupname)
	var gs = user_map[session.Name].GroupSort
	for i, v := range gs {
		if v == groupname {
			gs = append(gs[:i], gs[i+1:]...)
			user_map[session.Name].GroupSort = gs
			break
		}
	}
	cache_save(session.Name)
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

// 保存缓存
func cache_save(username string) {
	byte, _ := json.MarshalIndent(user_map[username], "", "    ")
	file, err := os.OpenFile(USERS_DIR+"/"+username+".json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	file.Write(byte)
}

// 会话保存
func session_save(session_id string) {
	byte, _ := json.MarshalIndent(session_map[session_id], "", "    ")
	file, err := os.OpenFile(SESSIONS_DIR+"/"+session_id+".json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	file.Write(byte)
}

// 会话加载
func session_load(session_file_name string) {
	content, err := os.ReadFile(SESSIONS_DIR + "/" + session_file_name)
	if err != nil {
		panic(err)
	}
	var session_info UserSession
	if nil != json.Unmarshal(content, &session_info) {
		panic("json parse error " + session_file_name)
	}
	session_map[strings.Split(session_file_name, ".")[0]] = &session_info
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
		if nil == user_map[username].Group {
			user_map[username].Group = make(map[string]*GroupInfo)
		}
		user_map[username].Group[group] = &GroupInfo{
			Name:      group,
			Markdowns: []string{},
		}
		user_map[username].GroupSort = append(user_map[username].GroupSort, group)
		// 刷盘
		cache_save(username)
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
	if r.URL.Path == "/export/" {
		// 导出整个用户的文档
		var fname = username + ".zip"
		w.Header().Add("Content-Disposition", "attachment; filename="+fname)
		w.Header().Add("Content-Type", "application/octet-stream")
		ZipDir(DATA_DIR+"/"+username, w)
	} else {
		var restname = strings.TrimPrefix(r.URL.Path, "/export/")
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
	parts := GetPathList(r.URL.Path, "/upload/")
	var groupname = parts[0]
	var markdownname = parts[1]

	var work_dir = DATA_DIR + "/" + session.Name + "/" + groupname + "/" + markdownname
	// 文件夹不存在就创建
	if _, err := os.Stat(work_dir); os.IsNotExist(err) {
		err := os.MkdirAll(work_dir, 0755)
		if err != nil {
			return
		}
	}

	// 从请求中获取文件
	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// 创建一个新文件
	newFile, err := os.Create(work_dir + "/" + header.Filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer newFile.Close()

	// 将上传的文件内容复制到新文件中
	_, err = io.Copy(newFile, file)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 返回上传成功的信息
	fmt.Fprintf(w, "文件上传成功")
}

func AuthError(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "index.html", http.StatusSeeOther)
}

func Auth(w http.ResponseWriter, r *http.Request) (bool, *UserSession) {
	// 需要权限控制的
	var cookie, err = r.Cookie("session_id")
	if err != nil {
		AuthError(w, r)
		return false, nil
	} else {
		var session_id = cookie.Value
		var session = session_map[session_id]
		// 校验session是否过期
		if session.Expires < time.Now().Unix() {
			delete(session_map, session_id)
			os.Remove(SESSIONS_DIR + "/" + session_id + ".json")
			return false, nil
		}
		if session == nil {
			// 用户未登录
			AuthError(w, r)
			return false, nil
		} else {
			// 用户已登录
			if strings.HasPrefix(r.URL.Path, "markdown/") {
				// 使用markdown下的资源需要判断权限
				if strings.HasPrefix(r.URL.Path, "markdown/"+session.Name+"/") {
					return true, session
				} else {
					// 无权限
					AuthError(w, r)
					return false, nil
				}
			} else {
				// 其余资源有权限
				return true, session
			}
		}

	}

}

func auth_markdown(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if suc, _ := Auth(w, r); suc {
			next.ServeHTTP(w, r)
		}
	})
}

// 从用户配置文件刷新缓存
func cache_load(user_file_name string) {
	content, err := os.ReadFile(USERS_DIR + "/" + user_file_name)
	if err != nil {
		panic(err)
	}
	var user_info UserInfo
	if nil != json.Unmarshal(content, &user_info) {
		panic("json parse error " + user_file_name)
	}
	user_map[strings.Split(user_file_name, ".")[0]] = &user_info
}

func init_work() {
	// 判断用户目录是否存在，不存在就创建默认的
	if _, err := os.Stat(USERS_DIR); os.IsNotExist(err) {
		err := os.MkdirAll(USERS_DIR, 0755)
		if err != nil {
			return
		}
	}
	f, err := os.Open(USERS_DIR)
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
	var usefiles []fs.FileInfo
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			usefiles = append(usefiles, file)
		}
	}
	if len(usefiles) == 0 {
		// 无任何用户配置，则初始化一个用户
		user_map["root"] = &UserInfo{
			Name:     "root",
			Password: "root",
		}
		fmt.Println("初始化用户名: root")
		fmt.Println("初始化密码: root")
		cache_save("root")
	}
	for _, file := range files {
		if !file.IsDir() {
			cache_load(file.Name())
		}
	}
	// 会话目录创建
	if _, err := os.Stat(SESSIONS_DIR); os.IsNotExist(err) {
		err := os.MkdirAll(SESSIONS_DIR, 0755)
		if err != nil {
			return
		}
	}
	// 加载会话
	fsd, err := os.Open(SESSIONS_DIR)
	if err != nil {
		fmt.Println("打开目录时出错：", err)
		return
	}
	defer fsd.Close()
	sfiles, err := fsd.Readdir(-1)
	if err != nil {
		fmt.Println("读取目录时出错：", err)
		return
	}
	var usesfiles []fs.FileInfo
	for _, file := range sfiles {
		if strings.HasSuffix(file.Name(), ".json") {
			usesfiles = append(usesfiles, file)
		}
	}
	if len(usesfiles) != 0 {
		// 加载会话
		for _, fi := range usesfiles {
			session_load(fi.Name())
		}
	}
}

// 定时任务
func Job() {
	var dur = 1 * time.Hour
	t := time.NewTimer(dur)
	for {
		<-t.C
		t.Reset(dur)
		// 把超时的session踢出去
		for k, us := range session_map {
			if us.Expires < time.Now().Unix() {
				// session 超时
				delete(session_map, k)
				os.Remove(SESSIONS_DIR + "/" + k + ".json")
			}
		}
	}
}

//go:embed page/*
var staticFiles embed.FS

func main() {
	var bind string
	flag.StringVar(&USERS_DIR, "users", "users", "用户信息存档目录")
	flag.StringVar(&DATA_DIR, "data", "markdown", "文档存储目录")
	flag.StringVar(&bind, "bind", "127.0.0.1:11990", "绑定host与端口信息")
	flag.StringVar(&SESSIONS_DIR, "sessions", "sessions", "会话持久化目录")
	flag.Parse()
	fmt.Println("浏览器地址：http://" + bind)
	init_work()
	go Job()
	webroot, _ := fs.Sub(staticFiles, "page")
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(webroot))))
	// http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("page/"))))
	http.Handle("/markdown/", http.StripPrefix("/markdown", auth_markdown(http.FileServer(http.Dir(DATA_DIR+"/")))))
	http.HandleFunc("/upload/", upload)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/group-list", group_list)
	http.HandleFunc("/new-group", new_group)
	http.HandleFunc("/new-markdown/", new_markdown)
	http.HandleFunc("/del-markdown/", del_markdown)
	http.HandleFunc("/del-group/", del_group)
	http.HandleFunc("/markdown-list", markdown_list)
	http.HandleFunc("/user-password-update", user_password_update)
	http.HandleFunc("/export/", export)
	server := http.Server{Addr: bind}
	server.ListenAndServe()
}
