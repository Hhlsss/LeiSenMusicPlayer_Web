package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

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
