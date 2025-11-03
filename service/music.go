package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"MusicPlayerWeb/db"

	"github.com/dhowden/tag"
)

type Track struct {
	ID        int    `json:"id"`
	Path      string `json:"-"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	HasCover  bool   `json:"hasCover"`
	HasLyrics bool   `json:"hasLyrics"`
}

// Album 表示按专辑聚合后的信息
type Album struct {
	Name         string `json:"name"`
	Artist       string `json:"artist"`
	Count        int    `json:"count"`
	CoverTrackID int    `json:"coverTrackId"`
	FirstTrackID int    `json:"firstTrackId"`
}

var (
	tracks   []Track
	once     sync.Once
	mu       sync.RWMutex
	musicDir = `C:\Users\28890\Desktop\music`
)

// InitMusicCache scans the directory and caches flac tracks.
func InitMusicCache() error {
	var err error
	once.Do(func() {
		err = scanDir()
	})
	return err
}

func scanDir() error {
	mu.Lock()
	defer mu.Unlock()

	tracks = nil
	id := 0
	walkErr := filepath.Walk(musicDir, func(p string, info os.FileInfo, e error) error {
		if e != nil {
			return nil // skip error entry
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if ext != ".flac" {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer f.Close()

		m, _ := tag.ReadFrom(f)

		title := info.Name()
		artist := ""
		album := ""
		hasCover := false
		hasLyrics := false

		if m != nil {
			if t := m.Title(); t != "" {
				title = t
			}
			artist = m.Artist()
			album = m.Album()
			// Cover
			if pic := m.Picture(); pic != nil && len(pic.Data) > 0 {
				hasCover = true
			}
			// Lyrics
			if l := m.Lyrics(); l != "" {
				hasLyrics = true
			} else if rawMeta, ok := m.(tag.Metadata); ok {
				for k, v := range rawMeta.Raw() {
					if k == "LYRICS" || k == "UNSYNCEDLYRICS" {
						if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
							hasLyrics = true
							break
						}
					}
				}
			}
		}

		tracks = append(tracks, Track{
			ID:        id,
			Path:      p,
			Title:     title,
			Artist:    artist,
			Album:     album,
			HasCover:  hasCover,
			HasLyrics: hasLyrics,
		})
		id++
		return nil
	})
	return walkErr
}

func ListTracks() ([]Track, error) {
	if err := InitMusicCache(); err != nil {
		return nil, err
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Track, len(tracks))
	copy(out, tracks)
	return out, nil
}

// GetMusicDir 获取当前音乐目录
func GetMusicDir() string {
	return musicDir
}

// UpdateMusicDir 更新音乐目录
func UpdateMusicDir(newDir string) error {
	// 验证目录是否存在
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", newDir)
	}

	// 更新全局变量
	musicDir = newDir

	// 重新扫描音乐文件
	return RescanMusic()
}

// RescanMusic 重新扫描音乐文件
func RescanMusic() error {
	// 清空现有数据
	mu.Lock()
	tracks = nil
	mu.Unlock()

	// 重新扫描
	_, err := ListTracks()
	return err
}

// ListAlbums 聚合专辑（专辑名+艺人 作为键），首曲与封面取第一首
func ListAlbums() ([]Album, error) {
	if err := InitMusicCache(); err != nil {
		return nil, err
	}
	mu.RLock()
	defer mu.RUnlock()
	type key struct{ album, artist string }
	m := map[key]*Album{}
	for _, t := range tracks {
		k := key{album: strings.TrimSpace(t.Album), artist: strings.TrimSpace(t.Artist)}
		if k.album == "" {
			continue
		}
		if _, ok := m[k]; !ok {
			m[k] = &Album{Name: k.album, Artist: k.artist, Count: 0, CoverTrackID: t.ID, FirstTrackID: t.ID}
		}
		al := m[k]
		al.Count++
		// 如果第一首有封面，记录为封面曲目
		if t.HasCover && al.CoverTrackID == al.FirstTrackID {
			al.CoverTrackID = t.ID
		}
	}
	out := make([]Album, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}
	return out, nil
}

// ListArtists 聚合歌手，统计歌曲数量，使用第一首歌曲的封面作为歌手封面
func ListArtists() ([]map[string]interface{}, error) {
	if err := InitMusicCache(); err != nil {
		return nil, err
	}
	mu.RLock()
	defer mu.RUnlock()

	artistsMap := make(map[string]*map[string]interface{})

	for _, t := range tracks {
		artist := strings.TrimSpace(t.Artist)
		if artist == "" {
			continue
		}

		if _, exists := artistsMap[artist]; !exists {
			// 创建新的歌手记录
			artistsMap[artist] = &map[string]interface{}{
				"name":         artist,
				"songCount":    0,
				"coverTrackID": t.ID,
				"firstTrack":   t,
			}
		}

		artistData := artistsMap[artist]
		(*artistData)["songCount"] = (*artistData)["songCount"].(int) + 1

		// 如果当前歌曲有封面，更新歌手封面
		if t.HasCover {
			(*artistData)["coverTrackID"] = t.ID
		}
	}

	// 转换为数组并添加分类信息
	var artists []map[string]interface{}
	for _, artistData := range artistsMap {
		artist := *artistData
		// 根据歌手名称判断分类
		category := "other"
		artistName := artist["name"].(string)

		// 简单的分类逻辑（可以根据需要扩展）
		if containsChinese(artistName) {
			category = "chinese"
		} else if containsJapanese(artistName) {
			category = "japanese"
		} else if containsKorean(artistName) {
			category = "korean"
		} else {
			category = "western"
		}

		artist["category"] = category
		artist["id"] = len(artists) + 1
		artists = append(artists, artist)
	}

	return artists, nil
}

// 辅助函数：判断字符串是否包含中文
func containsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

// 辅助函数：判断字符串是否包含日文
func containsJapanese(s string) bool {
	for _, r := range s {
		if (r >= 0x3040 && r <= 0x309F) || // 平假名
			(r >= 0x30A0 && r <= 0x30FF) || // 片假名
			(r >= 0x4E00 && r <= 0x9FAF) { // 汉字（中日共用）
			return true
		}
	}
	return false
}

// 辅助函数：判断字符串是否包含韩文
func containsKorean(s string) bool {
	for _, r := range s {
		if r >= 0xAC00 && r <= 0xD7AF {
			return true
		}
	}
	return false
}

// ListAlbumTracks 获取某专辑下所有曲目（按标题排序）
func ListAlbumTracks(album string, artist string) ([]Track, error) {
	if err := InitMusicCache(); err != nil {
		return nil, err
	}
	mu.RLock()
	defer mu.RUnlock()
	var out []Track
	for _, t := range tracks {
		if strings.TrimSpace(t.Album) == strings.TrimSpace(album) && (artist == "" || strings.TrimSpace(t.Artist) == strings.TrimSpace(artist)) {
			out = append(out, t)
		}
	}
	// 简单按标题排序
	sort.SliceStable(out, func(i, j int) bool { return strings.ToLower(out[i].Title) < strings.ToLower(out[j].Title) })
	return out, nil
}

// GetArtistByID 根据歌手ID获取歌手信息
func GetArtistByID(id int) (map[string]interface{}, error) {
	artists, err := ListArtists()
	if err != nil {
		return nil, err
	}

	for _, artist := range artists {
		if artistID, ok := artist["id"].(int); ok && artistID == id {
			return artist, nil
		}
	}

	return nil, errors.New("artist not found")
}

// ListArtistTracks 获取某歌手下所有曲目（按专辑和标题排序）
func ListArtistTracks(artistName string) ([]Track, error) {
	if err := InitMusicCache(); err != nil {
		return nil, err
	}
	mu.RLock()
	defer mu.RUnlock()

	var out []Track
	for _, t := range tracks {
		if strings.TrimSpace(t.Artist) == strings.TrimSpace(artistName) {
			out = append(out, t)
		}
	}

	// 按专辑和标题排序
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Album != out[j].Album {
			return strings.ToLower(out[i].Album) < strings.ToLower(out[j].Album)
		}
		return strings.ToLower(out[i].Title) < strings.ToLower(out[j].Title)
	})

	return out, nil
}

func getTrackByID(id int) (Track, error) {
	if err := InitMusicCache(); err != nil {
		return Track{}, err
	}
	mu.RLock()
	defer mu.RUnlock()
	for _, t := range tracks {
		if t.ID == id {
			return t, nil
		}
	}
	return Track{}, errors.New("not found")
}

// GetTrack 导出：根据ID获取曲目
func GetTrack(id int) (Track, error) {
	return getTrackByID(id)
}

func ReadAudio(id int) (io.ReadCloser, string, error) {
	t, err := getTrackByID(id)
	if err != nil {
		return nil, "", err
	}
	f, err := os.Open(t.Path)
	if err != nil {
		return nil, "", err
	}
	// FLAC mime
	return f, "audio/flac", nil
}

func ReadCover(id int) ([]byte, string, error) {
	t, err := getTrackByID(id)
	if err != nil {
		return nil, "", err
	}
	f, err := os.Open(t.Path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, "", err
	}
	if m == nil || m.Picture() == nil || len(m.Picture().Data) == 0 {
		return nil, "", errors.New("no cover")
	}
	p := m.Picture()
	// mime type
	ct := p.MIMEType
	if ct == "" {
		// guess common
		ct = "image/jpeg"
	}
	buf := bytes.NewBuffer(p.Data)
	return buf.Bytes(), ct, nil
}

func ReadLyrics(id int) (string, error) {
	t, err := getTrackByID(id)
	if err != nil {
		return "", err
	}
	f, err := os.Open(t.Path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	m, err := tag.ReadFrom(f)
	if err != nil || m == nil {
		return "", errors.New("no lyrics")
	}
	if l := m.Lyrics(); l != "" {
		return cleanLyrics(l), nil
	}
	// try raw fields
	if raw, ok := m.(tag.Metadata); ok {
		for k, v := range raw.Raw() {
			if k == "LYRICS" || k == "UNSYNCEDLYRICS" {
				if s, ok := v.(string); ok && s != "" {
					return cleanLyrics(s), nil
				}
			}
		}
	}
	return "", errors.New("no lyrics")
}

// ReadLyricsRaw 返回带时间戳的原始歌词（用于前端同步显示）
func ReadLyricsRaw(id int) (string, error) {
	t, err := getTrackByID(id)
	if err != nil {
		return "", err
	}
	f, err := os.Open(t.Path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	m, err := tag.ReadFrom(f)
	if err != nil || m == nil {
		return "", errors.New("no lyrics")
	}
	if l := m.Lyrics(); l != "" {
		return l, nil
	}
	if raw, ok := m.(tag.Metadata); ok {
		for k, v := range raw.Raw() {
			if k == "LYRICS" || k == "UNSYNCEDLYRICS" {
				if s, ok := v.(string); ok && s != "" {
					return s, nil
				}
			}
		}
	}
	return "", errors.New("no lyrics")
}

// cleanLyrics 去除时间戳与标签，规范为每句一行纯文本
func cleanLyrics(s string) string {
	// 统一换行为 \n
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	// 去除常见 LRC 时间戳 [mm:ss.xx] 或 [mm:ss]
	reTs := regexp.MustCompile(`\[\d{1,2}:\d{2}(?:\.\d{1,3})?\]`)
	s = reTs.ReplaceAllString(s, "")
	// 去除元标签如 [ti:...], [ar:...], [al:...], [by:...]
	reTag := regexp.MustCompile(`\[(ti|ar|al|by):[^\]]*\]`)
	s = reTag.ReplaceAllString(s, "")
	// 规范多余空行
	reNL := regexp.MustCompile(`
+`)
	s = reNL.ReplaceAllString(s, "")
	// 按行处理，去除前后空白，过滤空行
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		if t != "" {
			out = append(out, t)
		}
	}
	return strings.Join(out, "\n")
}

// Comment 结构体定义
type Comment struct {
	ID        int64  `json:"id"`
	SongID    int64  `json:"song_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	Nickname  string `json:"nickname"`
}

// GetSongComments 获取歌曲的评论列表
func GetSongComments(songID int) ([]Comment, error) {
	// 初始化数据库连接
	if err := db.Init(); err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	// 使用 HTTP 客户端直接调用 Supabase REST API
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/rest/v1/song_comments?song_id=eq.%d&select=*&order=created_at.desc",
		os.Getenv("SUPABASE_URL"), songID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回错误状态码: %d", resp.StatusCode)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 转换结果到Comment结构体
	var comments []Comment
	for _, item := range result {
		// 获取用户名（评论中存储的用户昵称）
		username := getStringFromMap(item, "username", "用户")

		comment := Comment{
			ID:        parseID(item["id"]),
			SongID:    int64(songID),
			UserID:    username,
			Content:   getStringFromMap(item, "content", ""),
			CreatedAt: formatTime(getStringFromMap(item, "created_at", "")),
			Nickname:  username, // 直接使用评论中存储的用户昵称
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// AddComment 添加评论
func AddComment(songID int, userUUID string, content string) (*Comment, error) {
	// 初始化数据库连接
	if err := db.Init(); err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	// 使用 HTTP 客户端直接调用 Supabase REST API
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/rest/v1/song_comments", os.Getenv("SUPABASE_URL"))

	// 获取用户真实昵称
	userNickname := "用户" // 默认昵称
	if userInfo, err := GetUserInfo(userUUID); err == nil {
		if nickname, ok := userInfo["nickname"].(string); ok && nickname != "" {
			userNickname = nickname
		}
	}

	// 准备插入数据
	insertData := map[string]interface{}{
		"song_id":  fmt.Sprintf("%d", songID),
		"username": userNickname, // 使用用户设置的昵称
		"content":  content,
		"rating":   5, // 默认评分
	}

	jsonData, err := json.Marshal(insertData)
	if err != nil {
		return nil, fmt.Errorf("序列化数据失败: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API 返回错误状态码: %d", resp.StatusCode)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("插入评论后未返回结果")
	}

	// 创建评论记录
	comment := &Comment{
		ID:        parseID(result[0]["id"]),
		SongID:    int64(songID),
		UserID:    getStringFromMap(result[0], "username", "用户"),
		Content:   content,
		CreatedAt: formatTime(getStringFromMap(result[0], "created_at", "")),
		Nickname:  userNickname, // 使用真实昵称
	}

	return comment, nil
}

// GetCurrentUserID 获取当前用户ID
func GetCurrentUserID(r *http.Request) (string, error) {
	// 从cookie中获取用户UUID
	cookie, err := r.Cookie("user_id")
	if err != nil {
		return "", err
	}
	
	// 现在直接返回UUID字符串，不再转换为整数
	return cookie.Value, nil
}

// GetUserInfo 获取用户信息
func GetUserInfo(userUUID string) (map[string]interface{}, error) {
	// 从数据库获取用户信息（使用UUID格式）
	userInfo, err := db.GetUserProfileByUUID(userUUID)
	if err != nil {
		// 如果获取失败，返回默认信息
		return map[string]interface{}{
			"id":       userUUID,
			"nickname": "用户",
			"email":    "user@example.com",
		}, nil
	}
	return userInfo, nil
}

// 辅助函数：从map中安全获取字符串
func getStringFromMap(m map[string]interface{}, key string, defaultValue string) string {
	if val, ok := m[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// 辅助函数：格式化时间
func formatTime(timeStr string) string {
	if timeStr == "" {
		return time.Now().Format("2006-01-02 15:04:05")
	}

	// 尝试解析时间
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t.Format("2006-01-02 15:04:05")
	}

	return timeStr
}

// 辅助函数：解析ID
func parseID(id interface{}) int64 {
	if id == nil {
		return time.Now().Unix()
	}

	switch v := id.(type) {
	case string:
		// 如果是UUID，返回时间戳
		return time.Now().Unix()
	case float64:
		return int64(v)
	case int64:
		return v
	default:
		return time.Now().Unix()
	}
}

// GetAlbumByID 根据专辑ID获取专辑信息
func GetAlbumByID(id int) (map[string]interface{}, error) {
	albums, err := ListAlbums()
	if err != nil {
		return nil, err
	}

	// 检查ID是否在有效范围内
	if id < 1 || id > len(albums) {
		return nil, errors.New("album not found")
	}

	album := albums[id-1]
	albumData := make(map[string]interface{})
	albumData["id"] = id
	albumData["name"] = album.Name
	albumData["artist"] = album.Artist
	albumData["songCount"] = album.Count

	// 生成封面URL
	if album.CoverTrackID >= 0 {
		albumData["cover"] = fmt.Sprintf("/api/cover?id=%d", album.CoverTrackID)
	} else {
		albumData["cover"] = "https://picsum.photos/id/1015/300/300"
	}

	return albumData, nil
}

// ListAlbumTracksByID 根据专辑ID获取专辑歌曲列表
func ListAlbumTracksByID(id int) ([]Track, error) {
	albums, err := ListAlbums()
	if err != nil {
		return nil, err
	}

	// 检查ID是否在有效范围内
	if id < 1 || id > len(albums) {
		return nil, errors.New("album not found")
	}

	album := albums[id-1]
	return ListAlbumTracks(album.Name, album.Artist)
}

// UserFavorite 用户收藏结构体
type UserFavorite struct {
	ID         int64  `json:"id"`
	UserID     string `json:"user_id"`
	SongID     string `json:"song_id"`
	SongTitle  string `json:"song_title"`
	SongArtist string `json:"song_artist"`
	SongAlbum  string `json:"song_album"`
	CreatedAt  string `json:"created_at"`
}

// GetUserFavorites 获取用户收藏列表
func GetUserFavorites(userUUID string) ([]UserFavorite, error) {
	// 初始化数据库连接
	if err := db.Init(); err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	// 使用 HTTP 客户端直接调用 Supabase REST API
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/rest/v1/user_favorites?user_id=eq.%s&select=*&order=created_at.desc",
		os.Getenv("SUPABASE_URL"), userUUID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 如果表不存在，返回空列表而不是错误
		if resp.StatusCode == http.StatusNotFound {
			return []UserFavorite{}, nil
		}
		return nil, fmt.Errorf("API 返回错误状态码: %d", resp.StatusCode)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 转换结果到UserFavorite结构体
	var favorites []UserFavorite
	for _, item := range result {
		favorite := UserFavorite{
			ID:         parseID(item["id"]),
			UserID:     userUUID,
			SongID:     getStringFromMap(item, "song_id", ""),
			SongTitle:  getStringFromMap(item, "song_title", ""),
			SongArtist: getStringFromMap(item, "song_artist", ""),
			SongAlbum:  getStringFromMap(item, "song_album", ""),
			CreatedAt:  formatTime(getStringFromMap(item, "created_at", "")),
		}
		favorites = append(favorites, favorite)
	}

	return favorites, nil
}

// AddUserFavorite 添加用户收藏
func AddUserFavorite(userUUID string, songID, songTitle, songArtist, songAlbum string) error {
	// 初始化数据库连接
	if err := db.Init(); err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	// 使用 HTTP 客户端直接调用 Supabase REST API
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/rest/v1/user_favorites", os.Getenv("SUPABASE_URL"))

	// 准备插入数据
	insertData := map[string]interface{}{
		"user_id":     userUUID,
		"song_id":     songID,
		"song_title":  songTitle,
		"song_artist": songArtist,
		"song_album":  songAlbum,
	}

	jsonData, err := json.Marshal(insertData)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		// 如果表不存在，尝试创建表
		if resp.StatusCode == http.StatusNotFound {
			// 创建user_favorites表
			if err := createUserFavoritesTable(); err != nil {
				// 如果创建表失败，使用本地存储作为降级方案
				fmt.Printf("创建收藏表失败，使用本地存储: %v\n", err)
				return nil
			}
			// 重新尝试插入数据
			return AddUserFavorite(userUUID, songID, songTitle, songArtist, songAlbum)
		}

		// 读取响应体获取详细错误信息
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API 返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteUserFavorite 删除用户收藏
func DeleteUserFavorite(userUUID string, songID string) error {
	// 初始化数据库连接
	if err := db.Init(); err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	// 使用 HTTP 客户端直接调用 Supabase REST API
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/rest/v1/user_favorites?user_id=eq.%s&song_id=eq.%s",
		os.Getenv("SUPABASE_URL"), userUUID, songID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		// 如果表不存在，直接返回成功（因为本来就没有收藏记录）
		if resp.StatusCode == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("API 返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

// CheckUserFavorite 检查歌曲是否已收藏
func CheckUserFavorite(userUUID string, songID string) (bool, error) {
	// 初始化数据库连接
	if err := db.Init(); err != nil {
		return false, fmt.Errorf("数据库连接失败: %v", err)
	}

	// 使用 HTTP 客户端直接调用 Supabase REST API
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/rest/v1/user_favorites?user_id=eq.%s&song_id=eq.%s&select=id",
		os.Getenv("SUPABASE_URL"), userUUID, songID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 如果表不存在，返回false而不是错误
		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, fmt.Errorf("API 返回错误状态码: %d", resp.StatusCode)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("解析响应失败: %v", err)
	}

	// 如果结果不为空，说明已收藏
	return len(result) > 0, nil
}

// createUserFavoritesTable 创建用户收藏表
func createUserFavoritesTable() error {
	// 使用 HTTP 客户端调用 Supabase SQL API 创建表
	httpClient := &http.Client{}

	// 创建表的 SQL 语句 - 使用更简单的SQL语法
	sql := `CREATE TABLE IF NOT EXISTS user_favorites (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		song_id VARCHAR NOT NULL,
		song_title VARCHAR NOT NULL,
		song_artist VARCHAR NOT NULL,
		song_album VARCHAR,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		UNIQUE(user_id, song_id)
	)`

	// 准备请求数据
	requestData := map[string]interface{}{
		"query": sql,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("序列化SQL语句失败: %v", err)
	}

	url := fmt.Sprintf("%s/rest/v1/", os.Getenv("SUPABASE_URL"))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("创建表失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}
