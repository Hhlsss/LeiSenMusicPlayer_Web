package controller

import (
	"net/http"
	"path/filepath"
)

// HandleIndex 首页：GET /
func HandleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "index.html"))
}

// HandleSong 歌曲页：GET /song
func HandleSong(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "song.html"))
}

// HandleArtists 歌手页：GET /artists
func HandleArtists(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "singer.html"))
}

// HandleAlbumsPage 专辑列表页：GET /albums
func HandleAlbumsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "albums.html"))
}

// HandleAlbumPage 专辑详情页：GET /album
func HandleAlbumPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "album.html"))
}

// HandleArtistPage 歌手详情页：GET /artist/{id}
func HandleArtistPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "artist.html"))
}

// HandleFavoritesPage 收藏页面：GET /favorites
func HandleFavoritesPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "favorites.html"))
}

// HandleProfilePage 个人中心页面：GET /profile
func HandleProfilePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "profile.html"))
}

// HandleLoginPage 登录页面：GET /login
func HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "login.html"))
}

// HandlePlaylistsPage 歌单页面：GET /playlists
func HandlePlaylistsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "playlists.html"))
}

// HandleRequirementsPage 需求文档页面：GET /requirements
func HandleRequirementsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "requirements.html"))
}
