package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"MusicPlayerWeb/service"
)

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// GET /api/music
func HandleMusicList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	list, err := service.ListTracks()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

// GET /api/albums -> 专辑聚合
func HandleAlbums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	list, err := service.ListAlbums()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 处理真实数据，添加分类和封面URL
	var result []map[string]interface{}
	for i, album := range list {
		albumData := make(map[string]interface{})
		albumData["id"] = i + 1
		albumData["name"] = album.Name
		albumData["artist"] = album.Artist
		albumData["songCount"] = album.Count

		// 根据歌手名称判断分类
		category := "other"
		if containsChinese(album.Artist) {
			category = "chinese"
		} else if containsJapanese(album.Artist) {
			category = "japanese"
		} else if containsKorean(album.Artist) {
			category = "korean"
		} else {
			category = "western"
		}
		albumData["category"] = category

		// 如果有封面曲目ID，生成封面URL
		if album.CoverTrackID >= 0 {
			albumData["cover"] = fmt.Sprintf("/api/cover?id=%d", album.CoverTrackID)
		} else {
			// 使用默认封面
			albumData["cover"] = "https://picsum.photos/id/1015/300/300"
		}

		result = append(result, albumData)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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

// GET /api/album_tracks?album=...&artist=
func HandleAlbumTracks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	album := r.URL.Query().Get("album")
	artist := r.URL.Query().Get("artist")
	if strings.TrimSpace(album) == "" {
		writeErr(w, http.StatusBadRequest, "album is required")
		return
	}
	list, err := service.ListAlbumTracks(album, artist)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

// GET /api/audio?id=...
func HandleAudio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	rc, ctype, err := service.ReadAudio(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	defer rc.Close()

	// 基础头
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Accept-Ranges", "bytes")

	// 支持 Range 请求（用于拖动进度）
	rangeHeader := r.Header.Get("Range")
	// 只有当底层是 *os.File 时才能精确 Seek 与计算长度
	if rangeHeader != "" {
		if f, ok := rc.(*os.File); ok {
			st, statErr := f.Stat()
			if statErr != nil {
				writeErr(w, http.StatusInternalServerError, statErr.Error())
				return
			}
			size := st.Size()

			// 解析 Range: bytes=start-end
			// 兼容 "bytes=start-" 或 "bytes=-suffix"
			if !strings.HasPrefix(rangeHeader, "bytes=") {
				// 非法 Range，返回整个文件
				w.WriteHeader(http.StatusOK)
				_, _ = io.Copy(w, f)
				return
			}
			spec := strings.TrimPrefix(rangeHeader, "bytes=")
			var start, end int64
			start = 0
			end = size - 1

			parts := strings.Split(spec, "-")
			if len(parts) == 2 {
				if parts[0] != "" {
					if s, perr := strconv.ParseInt(parts[0], 10, 64); perr == nil {
						start = s
					}
				} else {
					// 形如 "bytes=-N" 表示最后 N 字节
					if n, perr := strconv.ParseInt(parts[1], 10, 64); perr == nil {
						if n < size {
							start = size - n
						} else {
							start = 0
						}
					}
				}
				if parts[1] != "" {
					if e, perr := strconv.ParseInt(parts[1], 10, 64); perr == nil {
						end = e
					}
				}
			}

			// 边界校验
			if start < 0 || start >= size {
				// 无法满足
				w.Header().Set("Content-Range", "bytes */"+strconv.FormatInt(size, 10))
				w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				return
			}
			if end < start {
				end = size - 1
			}
			if end >= size {
				end = size - 1
			}

			// 定位并发送分段
			if _, seekErr := f.Seek(start, 0); seekErr != nil {
				writeErr(w, http.StatusInternalServerError, seekErr.Error())
				return
			}
			n := end - start + 1
			w.Header().Set("Content-Length", strconv.FormatInt(n, 10))
			w.Header().Set("Content-Range",
				"bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(size, 10))
			w.WriteHeader(http.StatusPartialContent)
			_, _ = io.CopyN(w, f, n)
			return
		}
	}

	// 无 Range 或不可 Seek：返回整个文件
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}

// GET /api/cover?id=...
func HandleCover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	data, ctype, err := service.ReadCover(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", ctype)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// GET /api/lyrics?id=...
func HandleLyrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	lyrics, err := service.ReadLyrics(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"lyrics": lyrics})
}

// GET /api/lyrics_raw?id=...
func HandleLyricsRaw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	lyrics, err := service.ReadLyricsRaw(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"lyrics": lyrics})
}

// GET /api/track?id=...
func HandleTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	t, err := service.GetTrack(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

// GET /api/artists
func HandleArtistsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	artists, err := service.ListArtists()
	if err != nil {
		// 如果扫描失败，返回示例数据
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"id":        1,
				"name":      "周杰伦",
				"category":  "chinese",
				"songCount": 156,
				"cover":     "https://picsum.photos/id/1015/300/300",
			},
			{
				"id":        2,
				"name":      "林俊杰",
				"category":  "chinese",
				"songCount": 89,
				"cover":     "https://picsum.photos/id/1016/300/300",
			},
			{
				"id":        3,
				"name":      "Taylor Swift",
				"category":  "western",
				"songCount": 234,
				"cover":     "https://picsum.photos/id/1018/300/300",
			},
			{
				"id":        4,
				"name":      "Ed Sheeran",
				"category":  "western",
				"songCount": 67,
				"cover":     "https://picsum.photos/id/1019/300/300",
			},
			{
				"id":        5,
				"name":      "BTS",
				"category":  "korean",
				"songCount": 123,
				"cover":     "https://picsum.photos/id/1020/300/300",
			},
			{
				"id":        6,
				"name":      "Blackpink",
				"category":  "korean",
				"songCount": 45,
				"cover":     "https://picsum.photos/id/1021/300/300",
			},
			{
				"id":        7,
				"name":      "Vaundy",
				"category":  "japanese",
				"songCount": 34,
				"cover":     "https://picsum.photos/id/1022/300/300",
			},
			{
				"id":        8,
				"name":      "Yoasobi",
				"category":  "japanese",
				"songCount": 28,
				"cover":     "https://picsum.photos/id/1023/300/300",
			},
		})
		return
	}

	// 处理真实数据，添加封面URL
	var result []map[string]interface{}
	for _, artist := range artists {
		artistData := make(map[string]interface{})
		artistData["id"] = artist["id"]
		artistData["name"] = artist["name"]
		artistData["category"] = artist["category"]
		artistData["songCount"] = artist["songCount"]

		// 如果有封面曲目ID，生成封面URL
		if coverTrackID, ok := artist["coverTrackID"].(int); ok {
			artistData["cover"] = fmt.Sprintf("/api/cover?id=%d", coverTrackID)
		} else {
			// 使用默认封面
			artistData["cover"] = "https://picsum.photos/id/1015/300/300"
		}

		result = append(result, artistData)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GET /api/artist_detail/{id}
func HandleArtistDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 从URL路径中提取歌手ID
	path := r.URL.Path
	idStr := strings.TrimPrefix(path, "/api/artist_detail/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid artist id")
		return
	}

	artist, err := service.GetArtistByID(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "artist not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artist)
}

// GET /api/artist_tracks/{id}
func HandleArtistTracks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 从URL路径中提取歌手ID
	path := r.URL.Path
	idStr := strings.TrimPrefix(path, "/api/artist_tracks/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid artist id")
		return
	}

	// 先获取歌手信息
	artist, err := service.GetArtistByID(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "artist not found")
		return
	}

	artistName, ok := artist["name"].(string)
	if !ok {
		writeErr(w, http.StatusInternalServerError, "invalid artist data")
		return
	}

	// 获取歌手的所有歌曲
	tracks, err := service.ListArtistTracks(artistName)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tracks)
}

// GET /api/get_music_dir
func HandleGetMusicDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"musicDir": service.GetMusicDir(),
	})
}

// POST /api/update_music_dir
func HandleUpdateMusicDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var request struct {
		MusicDir string `json:"musicDir"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if request.MusicDir == "" {
		writeErr(w, http.StatusBadRequest, "music directory cannot be empty")
		return
	}

	// 更新音乐目录
	if err := service.UpdateMusicDir(request.MusicDir); err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to update music directory")
		return
	}

	// 重新扫描音乐文件
	if err := service.RescanMusic(); err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to rescan music files")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Music directory updated successfully",
	})
}
