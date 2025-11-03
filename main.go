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

	// 创建自定义多路复用器
	mux := http.NewServeMux()

	// 静态资源：/static/* 映射到 ./web 目录
	fs := http.FileServer(http.Dir("web"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// 页面路由
	mux.HandleFunc("/", controller.HandleIndex)
	mux.HandleFunc("/login", controller.HandleLoginPage)
	mux.HandleFunc("/song", controller.HandleSong)
	mux.HandleFunc("/artists", controller.HandleArtists)
	mux.HandleFunc("/albums", controller.HandleAlbumsPage)
	mux.HandleFunc("/album", controller.HandleAlbumPage)
	mux.HandleFunc("/artist/", controller.HandleArtistPage)
	mux.HandleFunc("/favorites", controller.HandleFavoritesPage)
	mux.HandleFunc("/profile", controller.HandleProfilePage)
	mux.HandleFunc("/upload", controller.HandleUploadPage)
	mux.HandleFunc("/forum", controller.HandleForumPage)
	mux.HandleFunc("/forum/post/", controller.HandleForumPostPage)
	mux.HandleFunc("/playlists", controller.HandlePlaylistsPage)
	mux.HandleFunc("/requirements", controller.HandleRequirementsPage)

	// API 路由
	mux.HandleFunc("/api/register", controller.HandleRegister)
	mux.HandleFunc("/api/login", controller.HandleLogin)
	mux.HandleFunc("/api/logout", controller.HandleLogout)

	// 音乐数据 API
	mux.HandleFunc("/api/music", controller.HandleMusicList)
	mux.HandleFunc("/api/artists", controller.HandleArtistsAPI)
	mux.HandleFunc("/api/albums", controller.HandleAlbums)
	mux.HandleFunc("/api/album_tracks", controller.HandleAlbumTracks)
	mux.HandleFunc("/api/audio", controller.HandleAudio)
	mux.HandleFunc("/api/cover", controller.HandleCover)
	mux.HandleFunc("/api/lyrics", controller.HandleLyrics)
	mux.HandleFunc("/api/lyrics_raw", controller.HandleLyricsRaw)
	mux.HandleFunc("/api/track", controller.HandleTrack)
	mux.HandleFunc("/api/artist_detail/", controller.HandleArtistDetail)
	mux.HandleFunc("/api/artist_tracks/", controller.HandleArtistTracks)
	mux.HandleFunc("/api/get_music_dir", controller.HandleGetMusicDir)
	mux.HandleFunc("/api/update_music_dir", controller.HandleUpdateMusicDir)

	// 评论功能 API
	mux.HandleFunc("/api/comments", controller.HandleComments)
	mux.HandleFunc("/api/check_auth", controller.HandleCheckAuth)

	// 收藏功能 API
	mux.HandleFunc("/api/favorites", controller.HandleFavorites)
	mux.HandleFunc("/api/favorites/", controller.HandleFavoriteItem)
	mux.HandleFunc("/api/favorites/check", controller.HandleCheckFavorite)

	// 歌曲信息 API
	mux.HandleFunc("/api/song_info", controller.HandleSongInfo)
	mux.HandleFunc("/api/update_profile", controller.HandleUpdateProfile)

	// 音乐上传功能 API
	mux.HandleFunc("/api/upload/music", controller.HandleUploadMusic)
	mux.HandleFunc("/api/upload/status", controller.HandleUploadStatus)
	mux.HandleFunc("/api/upload/play", controller.HandlePlayUploadedMusic)
	mux.HandleFunc("/api/cloud/music", controller.HandleCloudMusicList)
	mux.HandleFunc("/api/cloud/stream", controller.HandleCloudMusicStream)

	// 论坛功能 API
	mux.HandleFunc("/api/forum/posts", controller.HandleForumPosts)
	mux.HandleFunc("/api/forum/post/", controller.HandleForumPost)
	mux.HandleFunc("/api/forum/replies", controller.HandleForumReplies)
	mux.HandleFunc("/api/forum/reply/", controller.HandleForumReply)
	mux.HandleFunc("/api/forum/stats", controller.HandleForumStats)
	mux.HandleFunc("/api/forum/my-posts", controller.HandleMyPosts)



	log.Println("Server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
