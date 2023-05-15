package main

import (
	"crypto/rand"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"
)

// 数据目录
var DATA = "markdown"

// 用户信息目录
var USERS = "users"

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
	LoginDate int64
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
				session_map[session_id] = &UserSession{
					SessionId: session_id,
					Name:      username,
					LoginDate: time.Now().Unix(),
				}
				cookie_session_id := http.Cookie{Name: "session_id", Value: session_id, Expires: time.Now().Add(30 * 24 * time.Hour)}
				http.SetCookie(w, &cookie_session_id)
				cookie_username := http.Cookie{Name: "username", Value: username, Expires: time.Now().Add(30 * 24 * time.Hour)}
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
	var fname = DATA + "/" + session.Name + "/" + groupname + "/" + markdownname + ".md"
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
	var fname = DATA + "/" + session.Name + "/" + groupname + "/" + markdownname
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
	var fname = DATA + "/" + session.Name + "/" + groupname
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
	var user_dir = DATA + "/" + name
	if _, err := os.Stat(user_dir); os.IsNotExist(err) {
		err := os.MkdirAll(user_dir, 0755)
		if err != nil {
			return
		}
	}
}

// 保存缓存
func cache_save(username string) {
	byte, _ := json.Marshal(user_map[username])
	file, err := os.OpenFile(USERS+"/"+username+".json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	file.Write(byte)
}

// 分组检测
func group_check(username string, group string) {
	user_check(username)
	var user_dir = DATA + "/" + username + "/" + group
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

	var work_dir = DATA + "/" + session.Name + "/" + groupname + "/" + markdownname
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
		var session = session_map[cookie.Value]
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
	content, err := os.ReadFile(USERS + "/" + user_file_name)
	if err != nil {
		panic(err)
	}
	var user_info UserInfo
	if nil != json.Unmarshal(content, &user_info) {
		panic("json parse error " + user_file_name)
	}
	user_map[strings.Split(user_file_name, ".")[0]] = &user_info
}

func init_user() {
	// 判断用户目录是否存在，不存在就创建默认的
	if _, err := os.Stat(USERS); os.IsNotExist(err) {
		err := os.MkdirAll(USERS, 0755)
		if err != nil {
			return
		}
	}
	f, err := os.Open(USERS)
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
}

//go:embed page/*
var staticFiles embed.FS

func main() {
	var bind string
	flag.StringVar(&USERS, "users", "users", "用户信息存档目录")
	flag.StringVar(&DATA, "data", "markdown", "文档存储目录")
	flag.StringVar(&bind, "bind", "127.0.0.1:11990", "绑定host与端口信息")
	flag.Parse()
	fmt.Println("浏览器地址：http://" + bind)
	init_user()
	webroot, _ := fs.Sub(staticFiles, "page")
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(webroot))))
	http.Handle("/markdown/", http.StripPrefix("/markdown", auth_markdown(http.FileServer(http.Dir(DATA+"/")))))
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
	server := http.Server{Addr: bind}
	server.ListenAndServe()
}
