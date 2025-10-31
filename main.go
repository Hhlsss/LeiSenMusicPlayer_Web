package main

import (
	"log"
	"net/http"

	"MusicPlayerWeb/controller"
	"MusicPlayerWeb/db"
)

func main() {
	// 初始化 Supabase
	if err := db.Init(); err != nil {
		log.Printf("Supabase 初始化失败: %v", err)
		log.Println("继续运行，但用户认证功能将不可用")
	}

	// 静态资源：/static/* 映射到 ./web 目录
	fs := http.FileServer(http.Dir("web"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 页面路由
	http.HandleFunc("/", controller.HandleIndex)
	http.HandleFunc("/song", controller.HandleSong)
	http.HandleFunc("/artists", controller.HandleArtists)
	http.HandleFunc("/albums", controller.HandleAlbumsPage)
	http.HandleFunc("/album", controller.HandleAlbumPage)
	http.HandleFunc("/artist/", controller.HandleArtistPage)
	http.HandleFunc("/favorites", controller.HandleFavoritesPage)
	http.HandleFunc("/profile", controller.HandleProfilePage)
	http.HandleFunc("/upload", controller.HandleUploadPage)

	// API 路由
	http.HandleFunc("/api/register", controller.HandleRegister)
	http.HandleFunc("/api/login", controller.HandleLogin)
	http.HandleFunc("/api/logout", controller.HandleLogout)

	// 音乐数据 API
	http.HandleFunc("/api/music", controller.HandleMusicList)
	http.HandleFunc("/api/artists", controller.HandleArtistsAPI)
	http.HandleFunc("/api/albums", controller.HandleAlbums)
	http.HandleFunc("/api/album_tracks", controller.HandleAlbumTracks)
	http.HandleFunc("/api/audio", controller.HandleAudio)
	http.HandleFunc("/api/cover", controller.HandleCover)
	http.HandleFunc("/api/lyrics", controller.HandleLyrics)
	http.HandleFunc("/api/lyrics_raw", controller.HandleLyricsRaw)
	http.HandleFunc("/api/track", controller.HandleTrack)
	http.HandleFunc("/api/artist_detail/", controller.HandleArtistDetail)
	http.HandleFunc("/api/artist_tracks/", controller.HandleArtistTracks)
	http.HandleFunc("/api/get_music_dir", controller.HandleGetMusicDir)
	http.HandleFunc("/api/update_music_dir", controller.HandleUpdateMusicDir)

	// 评论功能 API
	http.HandleFunc("/api/comments", controller.HandleComments)
	http.HandleFunc("/api/check_auth", controller.HandleCheckAuth)

	// 收藏功能 API
	http.HandleFunc("/api/favorites", controller.HandleFavorites)
	http.HandleFunc("/api/favorites/", controller.HandleFavoriteItem)
	http.HandleFunc("/api/favorites/check", controller.HandleCheckFavorite)

	// 歌曲信息 API
	http.HandleFunc("/api/song_info", controller.HandleSongInfo)
	http.HandleFunc("/api/update_profile", controller.HandleUpdateProfile)

	// 音乐上传功能 API
	http.HandleFunc("/api/upload/music", controller.HandleUploadMusic)
	http.HandleFunc("/api/upload/status", controller.HandleUploadStatus)

	log.Println("Server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
