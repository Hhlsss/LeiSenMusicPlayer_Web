package main

import (
	"log"
	"net/http"

	"MusicPlayerWeb/controller"
)

func main() {
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

	// API 路由
	http.HandleFunc("/api/register", controller.HandleRegister)
	http.HandleFunc("/api/login", controller.HandleLogin)

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

	log.Println("Server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
