// 全局登录状态管理
window.authManager = {
  // 检查登录状态
  async checkAuthStatus() {
    try {
      const response = await fetch('/api/check_auth');
      if (response.ok) {
        const data = await response.json();
        return data;
      }
    } catch (error) {
      console.error('检查登录状态失败:', error);
    }
    return { authenticated: false, user: null };
  },

  // 更新导航栏登录状态
  updateNavbarStatus(isLoggedIn, userInfo = null) {
    const loginBtn = document.getElementById('loginOpen');
    const registerBtn = document.getElementById('registerOpen');
    const logoutBtn = document.getElementById('logoutBtn');
    const nicknameEl = document.getElementById('userNickname');

    if (loginBtn) loginBtn.style.display = isLoggedIn ? 'none' : '';
    if (registerBtn) registerBtn.style.display = isLoggedIn ? 'none' : '';
    if (logoutBtn) {
      logoutBtn.style.display = isLoggedIn ? '' : 'none';
      logoutBtn.classList.toggle('hidden', !isLoggedIn);
    }
    
    if (nicknameEl) {
      if (isLoggedIn && userInfo) {
        // 使用数据库中的真实昵称，如果没有昵称则使用账号作为初始昵称
        const nickname = userInfo.nickname || userInfo.email || userInfo.account || '用户';
        nicknameEl.textContent = nickname;
        nicknameEl.style.display = '';
        nicknameEl.classList.remove('hidden');
      } else {
        nicknameEl.style.display = 'none';
        nicknameEl.classList.add('hidden');
      }
    }
  },

  // 初始化所有页面的登录状态
  async initAllPagesAuth() {
    const authData = await this.checkAuthStatus();
    this.updateNavbarStatus(authData.authenticated, authData.user);
    
    // 设置登录按钮跳转事件
    const loginBtn = document.getElementById('loginOpen');
    if (loginBtn) {
      loginBtn.addEventListener('click', () => {
        window.location.href = '/login.html';
      });
    }
    
    // 设置注册按钮跳转事件
    const registerBtn = document.getElementById('registerOpen');
    if (registerBtn) {
      registerBtn.addEventListener('click', () => {
        window.location.href = '/login.html';
      });
    }
    
    // 设置退出登录事件
    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) {
      logoutBtn.addEventListener('click', async () => {
        try {
          await fetch('/api/logout', { method: 'POST' });
          
          // 清除本地存储的用户信息
          localStorage.removeItem('currentUser');
          
          // 触发退出登录事件，通知所有页面更新状态
          window.dispatchEvent(new CustomEvent('userLogout'));
          
          // 更新导航栏状态
          window.authManager.updateNavbarStatus(false);
          
          // 显示登录/注册按钮
          const loginBtn = document.getElementById('loginOpen');
          const registerBtn = document.getElementById('registerOpen');
          if (loginBtn) loginBtn.style.display = '';
          if (registerBtn) registerBtn.style.display = '';
        } catch (error) {
          console.error('退出登录失败:', error);
        }
      });
    }
    
    // 监听用户登录事件
    window.addEventListener('userLogin', (event) => {
      this.updateNavbarStatus(true, event.detail);
    });
    
    // 监听用户信息更新事件
    window.addEventListener('userProfileUpdated', (event) => {
      this.updateNavbarStatus(true, event.detail);
    });
    
    // 监听用户退出登录事件
    window.addEventListener('userLogout', () => {
      this.updateNavbarStatus(false);
    });
  },

  // 获取当前用户信息
  getCurrentUser() {
    try {
      const userData = localStorage.getItem('currentUser');
      return userData ? JSON.parse(userData) : null;
    } catch (error) {
      console.error('获取当前用户信息失败:', error);
      return null;
    }
  },

  // 设置当前用户信息
  setCurrentUser(user) {
    try {
      localStorage.setItem('currentUser', JSON.stringify(user));
    } catch (error) {
      console.error('保存用户信息失败:', error);
    }
  }
};

/* 轮播初始化（支持动态更新） */
(function () {
  let initialized = false;
  window.initCarousel = function () {
    const slides = document.querySelectorAll(".slide");
    const prev = document.querySelector(".carousel-btn.prev");
    const next = document.querySelector(".carousel-btn.next");
    if (!slides.length || !prev || !next) return;
    let idx = 0;
    const show = (i) => {
      slides.forEach((s, k) => s.classList.toggle("active", k === i));
    };
    show(0);
    if (!initialized) {
      prev.addEventListener("click", () => { idx = (idx - 1 + slides.length) % slides.length; show(idx); });
      next.addEventListener("click", () => { idx = (idx + 1) % slides.length; show(idx); });
      setInterval(() => { idx = (idx + 1) % slides.length; show(idx); }, 5000);
      initialized = true;
    }
  };
  // 初次也尝试初始化一次（默认轮播图）
  window.initCarousel();
})();

// 首页热门推荐：改为从 /api/music 动态渲染本地 FLAC
(function () {
  const grid = document.querySelector(".grid");
  if (!grid) return;

  // 先尝试触发后端重扫，确保识别本地 FLAC（失败则忽略）
  fetch("/api/rescan").catch(() => {});
  fetch("/api/music")
    .then(res => res.json())
    .then(list => {
      if (!Array.isArray(list)) return;
      // 生成卡片
      grid.innerHTML = "";
      list.forEach(item => {
        const a = document.createElement("a");
        a.className = "card";
        a.href = "/song?id=" + encodeURIComponent(item.id);

        const thumb = document.createElement("div");
        thumb.className = "thumb";
        const coverUrl = item.hasCover ? ("/api/cover?id=" + item.id) : "https://picsum.photos/300/300";
        thumb.style.backgroundImage = "url('" + coverUrl + "')";
        a.appendChild(thumb);

        const info = document.createElement("div");
        info.className = "card-info";
        const h3 = document.createElement("h3");
        h3.textContent = item.title || "未知标题";
        const p = document.createElement("p");
        p.textContent = item.artist || "未知艺术家";
        const pc = document.createElement("span");
        pc.className = "playcount";
        pc.textContent = item.hasLyrics ? "含歌词" : "无歌词";
        info.appendChild(h3);
        info.appendChild(p);
        info.appendChild(pc);

        a.appendChild(info);
        grid.appendChild(a);
      });

      // 轮播图使用静态设置的图片，不再动态替换
    })
    .catch(() => {});
})();

// 云端音乐播放器类
class CloudMusicPlayer {
  constructor() {
    this.audio = null;
    this.currentTrack = null;
    this.playlist = [];
    this.currentIndex = -1;
    this.isCloudMusic = false;
    this.init();
  }

  init() {
    this.audio = document.getElementById("audio");
    if (!this.audio) {
      console.warn('Audio element not found');
      return;
    }

    // 监听音频加载事件
    this.audio.addEventListener('loadedmetadata', () => {
      this.updateDuration();
    });

    this.audio.addEventListener('timeupdate', () => {
      this.updateProgress();
    });

    this.audio.addEventListener('ended', () => {
      this.next();
    });
  }

  // 播放云端音乐
  async playCloudMusic(trackId, isCloud = true) {
    try {
      this.isCloudMusic = isCloud;
      
      // 获取音乐文件信息
      const response = await fetch(`/api/cloud/music`);
      const data = await response.json();
      
      // 合并云端和本地音乐
      this.playlist = [...data.cloud_music, ...data.local_music];
      
      // 查找当前曲目
      this.currentIndex = this.playlist.findIndex(track => track.id === trackId);
      if (this.currentIndex === -1) {
        throw new Error('Track not found');
      }

      this.currentTrack = this.playlist[this.currentIndex];
      
      // 设置音频源
      if (this.isCloudMusic) {
        // 云端音乐使用流媒体API
        this.audio.src = `/api/cloud/stream?id=${encodeURIComponent(trackId)}`;
      } else {
        // 本地音乐使用原有API
        this.audio.src = `/api/audio?id=${encodeURIComponent(trackId)}`;
      }

      // 更新UI
      this.updateTrackInfo();
      
      // 开始播放
      await this.audio.play();
      
      return true;
    } catch (error) {
      console.error('Failed to play cloud music:', error);
      return false;
    }
  }

  // 更新曲目信息
  updateTrackInfo() {
    if (!this.currentTrack) return;

    // 更新标题
    const titleEl = document.querySelector(".meta h1");
    if (titleEl) titleEl.textContent = this.currentTrack.title || "未知标题";

    // 更新艺术家
    const artistEl = document.querySelector(".meta .artist");
    if (artistEl) artistEl.textContent = "歌手：" + (this.currentTrack.artist || "未知艺术家");

    // 更新专辑
    const albumEl = document.querySelector(".meta .album");
    if (albumEl) albumEl.textContent = "所属专辑：" + (this.currentTrack.album || "未知专辑");

    // 更新封面
    const cover = document.querySelector(".disc-cover");
    if (cover) {
      if (this.isCloudMusic) {
        // 云端音乐使用默认封面
        cover.style.backgroundImage = "url('https://picsum.photos/300/300')";
      } else {
        cover.style.backgroundImage = "url('/api/cover?id=" + encodeURIComponent(this.currentTrack.id) + "')";
      }
    }

    // 更新底部播放器信息
    const miniCover = document.querySelector(".mini-cover");
    const trackTitle = document.querySelector(".track .title");
    const trackArtist = document.querySelector(".track .artist");
    
    if (miniCover) {
      if (this.isCloudMusic) {
        miniCover.src = 'https://picsum.photos/50/50';
      } else {
        miniCover.src = "/api/cover?id=" + encodeURIComponent(this.currentTrack.id);
      }
    }
    if (trackTitle) trackTitle.textContent = this.currentTrack.title || "未知标题";
    if (trackArtist) trackArtist.textContent = this.currentTrack.artist || "未知艺术家";
  }

  // 更新播放进度
  updateProgress() {
    const curTime = document.getElementById("curTime");
    const progress = document.getElementById("progress");
    
    if (curTime && this.audio) {
      const minutes = Math.floor(this.audio.currentTime / 60);
      const seconds = Math.floor(this.audio.currentTime % 60);
      curTime.textContent = `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    }
    
    if (progress && this.audio.duration) {
      const progressPercent = (this.audio.currentTime / this.audio.duration) * 100;
      progress.value = this.audio.currentTime;
      progress.style.setProperty('--progress', progressPercent + '%');
    }
  }

  // 更新总时长
  updateDuration() {
    const durTime = document.getElementById("durTime");
    const progress = document.getElementById("progress");
    
    if (durTime && this.audio.duration) {
      const minutes = Math.floor(this.audio.duration / 60);
      const seconds = Math.floor(this.audio.duration % 60);
      durTime.textContent = `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    }
    
    if (progress && this.audio.duration) {
      progress.max = this.audio.duration;
    }
  }

  // 播放/暂停
  togglePlay() {
    if (!this.audio) return;
    
    if (this.audio.paused) {
      this.audio.play();
    } else {
      this.audio.pause();
    }
  }

  // 下一首
  next() {
    if (this.playlist.length === 0) return;
    
    this.currentIndex = (this.currentIndex + 1) % this.playlist.length;
    this.currentTrack = this.playlist[this.currentIndex];
    
    if (this.isCloudMusic) {
      this.audio.src = `/api/cloud/stream?id=${encodeURIComponent(this.currentTrack.id)}`;
    } else {
      this.audio.src = `/api/audio?id=${encodeURIComponent(this.currentTrack.id)}`;
    }
    
    this.updateTrackInfo();
    this.audio.play();
  }

  // 上一首
  prev() {
    if (this.playlist.length === 0) return;
    
    this.currentIndex = (this.currentIndex - 1 + this.playlist.length) % this.playlist.length;
    this.currentTrack = this.playlist[this.currentIndex];
    
    if (this.isCloudMusic) {
      this.audio.src = `/api/cloud/stream?id=${encodeURIComponent(this.currentTrack.id)}`;
    } else {
      this.audio.src = `/api/audio?id=${encodeURIComponent(this.currentTrack.id)}`;
    }
    
    this.updateTrackInfo();
    this.audio.play();
  }

  // 设置音量
  setVolume(volume) {
    if (!this.audio) return;
    this.audio.volume = Math.max(0, Math.min(1, volume));
  }

  // 跳转到指定时间
  seekTo(time) {
    if (!this.audio || !this.audio.duration) return;
    this.audio.currentTime = Math.max(0, Math.min(this.audio.duration, time));
  }
}

// 初始化云端音乐播放器
window.cloudMusicPlayer = new CloudMusicPlayer();

// 歌曲页：根据 id 加载音频、封面与歌词
(function () {
  const url = new URL(window.location.href);
  const id = url.pathname.endsWith("/song") ? url.searchParams.get("id") : null;
  const audio = document.getElementById("audio");
  if (!id || !audio) return;

  // 检查是否为云端音乐
  const isCloudMusic = url.searchParams.get("source") === "cloud";
  
  if (isCloudMusic && window.cloudMusicPlayer) {
    // 使用云端播放器播放云端音乐
    window.cloudMusicPlayer.playCloudMusic(id, true);
  } else {
    // 使用原有方式播放本地音乐
    audio.src = "/api/audio?id=" + encodeURIComponent(id);

    // 等待音频加载完成（注意：歌词加载已由 song.html 中的 loadLyrics() 处理，这里不再重复加载）
    audio.addEventListener('loadedmetadata', function() {
      // 加载封面图片
      const cover = document.querySelector(".disc-cover");
      if (cover) {
        cover.style.backgroundImage = "url('/api/cover?id=" + encodeURIComponent(id) + "')";
      }

  // 加载曲目信息并更新歌名、歌手、专辑以及底部栏
  fetch("/api/track?id=" + encodeURIComponent(id))
    .then(res => res.json())
    .then(t => {
      const titleEl = document.querySelector(".meta h1");
      const artistEl = document.querySelector(".meta .artist");
      const albumEl = document.querySelector(".meta .album");
      if (titleEl) titleEl.textContent = t.title || "未知标题";
      if (artistEl) artistEl.textContent = "歌手：" + (t.artist || "未知艺术家");
      if (albumEl) albumEl.textContent = "所属专辑：" + (t.album || "未知专辑");

      // 底部播放器信息与封面
      const miniCover = document.querySelector(".mini-cover");
      if (miniCover) miniCover.src = "/api/cover?id=" + encodeURIComponent(id);
      const trackTitle = document.querySelector(".track .title");
      const trackArtist = document.querySelector(".track .artist");
      if (trackTitle) trackTitle.textContent = t.title || "未知标题";
      if (trackArtist) trackArtist.textContent = t.artist || "未知艺术家";
    })
    .catch(() => {})

      // 加载歌词（参考参考项目的实现）
      // 注意：song.html 中有自己的歌词加载逻辑，这里跳过避免冲突
      // 检查是否有 song.html 特有的元素来避免重复加载
      if (document.getElementById('favoriteBtn') || document.querySelector('.song-page')) {
        // 在 song.html 页面，歌词由页面自己的逻辑加载
        return;
      }
      
      const lyricsEl = document.querySelector(".lyrics");
      if (lyricsEl) {
        fetch("/api/lyrics?id=" + encodeURIComponent(id))
          .then(res => res.json())
          .then(j => {
        if (!j || !j.lyrics) { 
          lyricsEl.innerHTML = '<h3>歌词</h3><p>暂无歌词</p>'; 
          return; 
        }
        
        const lyricsText = j.lyrics;
        let parsedLines = []; // 用于存储解析后的歌词行（LRC格式）
        let hasLrcFormat = false;
        
        // 改进的LRC格式检测：支持时间戳在行首或行中的任何位置
        const lrcTimeRegex = /\[(\d{1,2}):(\d{2})(?:\.(\d{1,3}))?\]/g;
        hasLrcFormat = lrcTimeRegex.test(lyricsText);
        
        if (hasLrcFormat) {
          // 解析LRC格式歌词
          const lines = [];
          lyricsText.split(/\r?\n/).forEach(raw => {
            if (!raw.trim()) return;
            
            // 提取所有时间戳
            const timeMatches = [];
            let match;
            lrcTimeRegex.lastIndex = 0; // 重置正则
            while ((match = lrcTimeRegex.exec(raw)) !== null) {
              const min = parseInt(match[1], 10);
              const sec = parseInt(match[2], 10);
              const ms = match[3] ? parseInt(match[3], 10) : 0;
              
              let totalSeconds = min * 60 + sec;
              if (match[3]) {
                // 根据毫秒位数处理
                if (match[3].length === 3) {
                  totalSeconds += ms / 1000;
                } else if (match[3].length === 2) {
                  totalSeconds += ms / 100;
                } else {
                  totalSeconds += ms / 10;
                }
              }
              timeMatches.push(totalSeconds);
            }
            
            // 移除所有时间戳，获取歌词文本
            const textRegex = /\[(\d{1,2}):(\d{2})(?:\.(\d{1,3}))?\]/g;
            const text = raw.replace(textRegex, '').trim();
            
            if (timeMatches.length > 0 && text) {
              // 为每个时间戳创建一个条目
              timeMatches.forEach(time => {
                lines.push({ t: time, text: text });
              });
            } else if (text) {
              // 有文本但没有时间戳，保留原文本（可能是标签或注释）
              console.debug('跳过非时间戳行:', text);
            }
          });
          
          // 按时间排序并去重
          lines.sort((a, b) => a.t - b.t);
          const seen = new Set();
          lines.forEach(item => {
            const key = item.t.toFixed(3);
            if (!seen.has(key)) {
              seen.add(key);
              parsedLines.push(item);
            }
          });
          
          // 渲染LRC格式歌词
          if (parsedLines.length > 0) {
            lyricsEl.innerHTML = '<h3>歌词</h3>';
            parsedLines.forEach(ln => {
              const p = document.createElement("p");
              p.textContent = ln.text;
              p.dataset.t = ln.t;
              lyricsEl.appendChild(p);
            });
          } else {
            // LRC解析失败，回退到普通分行
            hasLrcFormat = false;
            renderPlainLyrics(lyricsEl, lyricsText);
          }
        } else {
          // 普通歌词，按行分割显示
          renderPlainLyrics(lyricsEl, lyricsText);
        }
        
        // 渲染普通歌词的函数
        function renderPlainLyrics(container, text) {
          container.innerHTML = '<h3>歌词</h3>';
          const lyricsLines = text.split(/\r?\n/).filter(line => line.trim());
          if (lyricsLines.length > 0) {
            lyricsLines.forEach(line => {
              const p = document.createElement("p");
              p.textContent = line.trim();
              container.appendChild(p);
            });
          } else {
            // 如果没有换行符，尝试其他分隔符
            if (text.includes('。') || text.includes('！') || text.includes('？')) {
              const segments = text.split(/[。！？]/).filter(s => s.trim());
              segments.forEach(seg => {
                const p = document.createElement("p");
                p.textContent = seg.trim();
                container.appendChild(p);
              });
            } else {
              container.innerHTML = '<h3>歌词</h3><p>暂无歌词</p>';
            }
          }
        }
        // 同步高亮和滚动
        let isUserScrolling = false;
        let scrollTimeout = null;
        
        // 监听用户滚动
        lyricsEl.addEventListener('scroll', () => {
          isUserScrolling = true;
          
          // 清除之前的定时器
          if (scrollTimeout) clearTimeout(scrollTimeout);
          
          // 设置2秒后自动回到高亮歌词
          scrollTimeout = setTimeout(() => {
            isUserScrolling = false;
          }, 2000);
        });
        
        // 点击歌词跳转播放功能（增强版）
        lyricsEl.addEventListener('click', (e) => {
          // 查找点击的元素（可能是 <p> 标签或其子元素）
          let targetElement = e.target;
          
          // 如果点击的不是 <p> 标签，向上查找父元素
          while (targetElement && targetElement.tagName !== 'P' && targetElement !== lyricsEl) {
            targetElement = targetElement.parentElement;
          }
          
          // 确保点击的是歌词行（<p> 标签）
          if (targetElement && targetElement.tagName === 'P') {
            const targetTime = parseFloat(targetElement.dataset.t || targetElement.getAttribute('data-time'));
            
            // 如果有时间戳，跳转到对应位置
            if (!isNaN(targetTime) && targetTime >= 0 && audio) {
              // 设置播放位置
              audio.currentTime = targetTime;
              
              // 如果音频暂停，自动播放
              if (audio.paused) {
                audio.play().catch(err => {
                  console.warn('自动播放失败:', err);
                });
              }
              
              // 添加点击反馈效果
              targetElement.style.transform = 'scale(1.05)';
              targetElement.style.transition = 'transform 0.2s';
              
              setTimeout(() => {
                targetElement.style.transform = '';
              }, 200);
              
              console.log('跳转到时间:', targetTime, '秒');
            } else {
              console.warn('该歌词行没有有效的时间戳');
            }
          }
        });
        
        // 设置歌词高亮和滚动（支持LRC和普通歌词）
        // 将 parsedLines 和 hasLrcFormat 保存到作用域变量中，以便在高亮函数中使用
        const lyricsData = {
          lines: parsedLines,
          hasLrcFormat: hasLrcFormat
        };
        
        const setupLyricsHighlight = () => {
          const children = Array.from(lyricsEl.children).filter(el => el.tagName === 'P' && el.parentElement === lyricsEl);
          if (children.length === 0) return;
          
          const highlight = (cur) => {
            let idx = -1;
            
            if (lyricsData.lines.length > 0 && lyricsData.hasLrcFormat) {
              // LRC格式：使用时间戳匹配
              for (let i = 0; i < lyricsData.lines.length; i++) {
                if (cur >= lyricsData.lines[i].t) idx = i; else break;
              }
              if (idx >= children.length) idx = children.length - 1;
            } else {
              // 普通歌词：根据播放进度估算行数
              // 假设每行歌词播放约5秒（可根据实际情况调整）
              const secondsPerLine = audio.duration ? (audio.duration / children.length) : 5;
              idx = Math.floor(cur / secondsPerLine);
              if (idx >= children.length) idx = children.length - 1;
            }
            
            // 移除所有高亮
            children.forEach(el => el.classList.remove("active"));
            
            // 高亮当前行
            if (idx >= 0 && idx < children.length) {
              children[idx].classList.add("active");
              
              // 如果用户没有在滚动，自动滚动到高亮歌词
              if (!isUserScrolling && children[idx]) {
                const activeElement = children[idx];
                const containerHeight = lyricsEl.clientHeight;
                const elementHeight = activeElement.offsetHeight;
                const elementTop = activeElement.offsetTop;
                
                // 计算目标滚动位置，使高亮歌词显示在中央
                const targetScrollTop = elementTop - (containerHeight / 2) + (elementHeight / 2);
                
                // 确保滚动位置在合理范围内
                const maxScrollTop = lyricsEl.scrollHeight - containerHeight;
                const adjustedScrollTop = Math.max(0, Math.min(targetScrollTop, maxScrollTop));
                
                // 平滑滚动到目标位置
                lyricsEl.scrollTo({
                  top: adjustedScrollTop,
                  behavior: 'smooth'
                });
              }
            }
          };
          
          // 监听音频时间更新（避免重复添加监听器）
          if (!audio.hasLyricsListener) {
            audio.addEventListener("timeupdate", () => {
              if (audio && lyricsEl) {
                highlight(audio.currentTime);
              }
            });
            audio.hasLyricsListener = true;
          }
          
          // 初始高亮
          highlight(0);
        };
        
        // 设置歌词高亮（在歌词渲染完成后）
        setTimeout(setupLyricsHighlight, 100);
      })
      .catch(() => { 
        lyricsEl.innerHTML = '<h3>歌词</h3><p>暂无歌词</p>'; 
      });
    }
  }); // 结束 audio.addEventListener('loadedmetadata')
}
};

// 播放器控制（歌曲页）
(function () {
  const audio = document.getElementById("audio");
  if (!audio) return;
  const toggleBtn = document.getElementById("toggleBtn");
  const playBtn = document.getElementById("playBtn");
  const prevBtn = document.getElementById("prevBtn");
  const nextBtn = document.getElementById("nextBtn");
  const progress = document.getElementById("progress");
  const volume = document.getElementById("volume");
  const curTime = document.getElementById("curTime");
  const durTime = document.getElementById("durTime");

  // 播放时封面旋转和按钮状态更新
  const discCover = document.querySelector(".disc-cover");
  const updatePlayButtonState = () => {
    if (toggleBtn) {
      if (audio.paused) {
        toggleBtn.classList.remove("playing");
        toggleBtn.textContent = "⏯";
      } else {
        toggleBtn.classList.add("playing");
        toggleBtn.textContent = "⏸";
      }
    }
  };
  
  if (discCover) {
    audio.addEventListener("play",   () => {
      discCover.classList.add("rotating");
      updatePlayButtonState();
    });
    audio.addEventListener("pause",  () => {
      discCover.classList.remove("rotating");
      updatePlayButtonState();
    });
    audio.addEventListener("ended",  () => {
      discCover.classList.remove("rotating");
      updatePlayButtonState();
    });
  }
  
  // 初始化按钮状态
  updatePlayButtonState();

  // 预取播放列表并定位当前索引
  let playlist = [];
  let curIndex = -1;
  (function preloadPlaylist() {
    fetch("/api/music")
      .then(r => r.json())
      .then(list => {
        if (!Array.isArray(list)) return;
        playlist = list;
        const url = new URL(window.location.href);
        const id = url.searchParams.get("id");
        curIndex = playlist.findIndex(x => String(x.id) === String(id));
      })
      .catch(() => {});
  })();

  const fmt = (s) => {
    const m = Math.floor(s / 60);
    const sec = Math.floor(s % 60);
    return String(m).padStart(2, "0") + ":" + String(sec).padStart(2, "0");
  };

  function updateDur() {
    if (isFinite(audio.duration)) {
      durTime.textContent = fmt(audio.duration);
    }
  }
  audio.addEventListener("loadedmetadata", () => {
    updateDur();
    if (isFinite(audio.duration)) {
      progress.max = Math.floor(audio.duration);
      progress.step = 1;
    }
  });
  audio.addEventListener("durationchange", () => {
    updateDur();
    if (isFinite(audio.duration)) {
      progress.max = Math.floor(audio.duration);
      progress.step = 1;
    }
  });

  // 更新进度条样式
  const updateProgressStyle = () => {
    if (isFinite(audio.duration) && audio.duration > 0) {
      const progressPercent = (audio.currentTime / audio.duration) * 100;
      progress.style.setProperty('--progress', progressPercent + '%');
    }
  };

  // 更新音量条样式
  const updateVolumeStyle = () => {
    const volumePercent = audio.volume * 100;
    volume.style.setProperty('--volume', volumePercent + '%');
    
    // 更新音量图标
    const volumeIcon = document.getElementById('volumeIcon');
    if (volumeIcon) {
      if (audio.volume === 0) {
        volumeIcon.textContent = '🔇';
      } else if (audio.volume < 0.3) {
        volumeIcon.textContent = '🔈';
      } else if (audio.volume < 0.7) {
        volumeIcon.textContent = '🔉';
      } else {
        volumeIcon.textContent = '🔊';
      }
    }
  };

  // 音量图标点击静音/取消静音
  const volumeIcon = document.getElementById('volumeIcon');
  if (volumeIcon) {
    volumeIcon.addEventListener('click', () => {
      if (audio.volume > 0) {
        audio.volume = 0;
      } else {
        audio.volume = 0.8; // 恢复默认音量
      }
      updateVolumeStyle();
    });
  }

  audio.addEventListener("timeupdate", () => {
    curTime.textContent = fmt(audio.currentTime);
    if (!isSeeking && isFinite(audio.duration)) {
      progress.value = Math.floor(audio.currentTime);
      updateProgressStyle();
    }
  });

  toggleBtn.addEventListener("click", async () => {
    if (audio.paused) await audio.play(); else audio.pause();
  });
  playBtn?.addEventListener("click", async () => { if (audio.paused) await audio.play(); });

  let isSeeking = false;
  const beginSeek = () => { isSeeking = true; };
  const endSeek = () => {
    isSeeking = false;
    if (isFinite(audio.duration)) {
      const v = Number(progress.value) || 0;
      audio.currentTime = v;
    }
  };

  // 鼠标与指针事件
  progress.addEventListener("mousedown", beginSeek);
  progress.addEventListener("mouseup", endSeek);
  progress.addEventListener("pointerdown", beginSeek);
  progress.addEventListener("pointerup", endSeek);

  // 触控事件（移动端）
  progress.addEventListener("touchstart", beginSeek, { passive: true });
  progress.addEventListener("touchend", endSeek, { passive: true });

  // 拖动过程中实时定位到对应秒数
  progress.addEventListener("input", () => {
    if (isFinite(audio.duration)) {
      const v = Number(progress.value) || 0;
      audio.currentTime = v;
    }
  });

  // 松手（change）时再次对齐
  progress.addEventListener("change", endSeek);
  // 在 timeupdate 时仅更新进度，不重置拖动中的位置（逻辑保持在上方的 timeupdate 中）

  volume.addEventListener("input", () => {
    audio.volume = volume.value / 100;
    updateVolumeStyle();
  });

  // 初始化样式
  updateVolumeStyle();
  updateProgressStyle();

  // 上一首/下一首：跳转到相邻歌曲详情页
  function gotoByOffset(off) {
    if (!playlist.length || curIndex < 0) return;
    const n = (curIndex + off + playlist.length) % playlist.length;
    const target = playlist[n];
    if (target) {
      window.location.href = "/song?id=" + encodeURIComponent(target.id);
    }
  }
  prevBtn.addEventListener("click", () => gotoByOffset(-1));
  nextBtn.addEventListener("click", () => gotoByOffset(1));
})();

// Supabase 初始化与会话监听
(function () {
  try {
    // 检查是否在支持 Supabase 的页面
    const loginBtn = document.getElementById('loginOpen');
    if (!loginBtn) return; // 非登录页面无需初始化
    
    const url = window.__SUPABASE_URL__;
    const key = window.__SUPABASE_ANON_KEY__;
    
    if (!url || !key || url === 'YOUR_SUPABASE_URL' || key === 'YOUR_SUPABASE_ANON_KEY') {
      console.warn('Supabase 未配置，跳过初始化');
      return;
    }
    
    // 确保 supabase 库已加载
    if (typeof window.supabase === 'undefined') {
      console.error('Supabase JS 库未加载');
      return;
    }
    
    const client = window.supabase.createClient(url, key);
    window.supabaseClient = client;

    const regBtn = document.getElementById('registerOpen');
    const logoutBtn = document.getElementById('logoutBtn');
    const nickEl = document.getElementById('userNickname');

    function applySignedIn(session) {
      if (loginBtn) loginBtn.style.display = 'none';
      if (regBtn) regBtn.style.display = 'none';
      if (logoutBtn) logoutBtn.classList.remove('hidden');
      const email = session?.user?.email || session?.user?.phone || '已登录';
      if (nickEl) { nickEl.textContent = email; nickEl.classList.remove('hidden'); }
    }
    function applySignedOut() {
      if (loginBtn) loginBtn.style.display = '';
      if (regBtn) regBtn.style.display = '';
      if (logoutBtn) logoutBtn.classList.add('hidden');
      if (nickEl) { nickEl.textContent = ''; nickEl.classList.add('hidden'); }
    }

    client.auth.getSession().then(({ data }) => {
      if (data?.session) applySignedIn(data.session); else applySignedOut();
    });

    client.auth.onAuthStateChange((event, session) => {
      if (event === 'SIGNED_IN' || event === 'TOKEN_REFRESHED' || event === 'USER_UPDATED' || event === 'INITIAL_SESSION') {
        applySignedIn(session);
      } else if (event === 'SIGNED_OUT') {
        applySignedOut();
      }
    });

    logoutBtn?.addEventListener('click', async () => {
      await client.auth.signOut();
      applySignedOut();
    });
  } catch (e) {
    console.error('Init Supabase failed', e);
  }
})();

// 登录/注册模态交互
function initModalInteractions() {
  const overlay = document.getElementById("modalOverlay");
  const loginModal = document.getElementById("loginModal");
  const registerModal = document.getElementById("registerModal");

  // 检查必要的元素是否存在
  if (!overlay || !loginModal || !registerModal) {
    console.warn('登录注册模态框元素未找到，请检查HTML结构');
    return;
  }

  const openModal = (modal) => {
    overlay.classList.remove("hidden");
    modal.classList.remove("hidden");
  };
  const closeModal = (modal) => {
    overlay.classList.add("hidden");
    modal.classList.add("hidden");
  };

  // 暴露模态框函数到全局作用域
  window.openLoginModal = () => openModal(loginModal);
  window.openRegisterModal = () => openModal(registerModal);
  window.closeLoginModal = () => closeModal(loginModal);
  window.closeRegisterModal = () => closeModal(registerModal);

  // 绑定登录按钮事件 - 跳转到登录页面
  const loginBtn = document.getElementById("loginOpen");
  if (loginBtn) {
    loginBtn.addEventListener("click", (e) => {
      e.preventDefault();
      window.location.href = '/login';
    });
  } else {
    console.warn('登录按钮未找到');
  }

  // 绑定注册按钮事件 - 跳转到登录页面（带注册参数）
  const registerBtn = document.getElementById("registerOpen");
  if (registerBtn) {
    registerBtn.addEventListener("click", (e) => {
      e.preventDefault();
      window.location.href = '/login?action=register';
    });
  } else {
    console.warn('注册按钮未找到');
  }

  // 关闭按钮
  document.querySelectorAll(".modal-close").forEach(btn => {
    btn.addEventListener("click", () => {
      const id = btn.getAttribute("data-close");
      const modal = document.getElementById(id);
      closeModal(modal);
    });
  });

  // 点击遮罩关闭
  overlay?.addEventListener("click", () => {
    closeModal(loginModal);
    closeModal(registerModal);
  });

  // Esc 关闭
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape") {
      closeModal(loginModal);
      closeModal(registerModal);
    }
  });

  // 使用自定义账号登录（邮箱/密码）
  document.getElementById("loginForm")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    const account = document.getElementById("loginAccount").value.trim();
    const pwd = document.getElementById("loginPassword").value;
    if (!account || !pwd) { alert("请输入账号和密码"); return; }

    const submitBtn = document.querySelector("#loginForm button[type='submit']");
    if (submitBtn) { submitBtn.disabled = true; submitBtn.textContent = "登录中…"; }

    try {
      const response = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ account, password: pwd })
      });
      
      if (response.ok) {
        const data = await response.json();
        alert("登录成功");
        closeModal(loginModal);
        
        // 更新本地存储的用户信息
        window.authManager.setCurrentUser({
          id: data.user_id,
          nickname: data.nickname,
          email: account
        });
        
        // 触发登录事件，通知所有页面更新状态
        window.dispatchEvent(new CustomEvent('userLogin', { detail: data }));
        
        // 更新导航栏状态
        window.authManager.updateNavbarStatus(true, {
          id: data.user_id,
          nickname: data.nickname,
          email: account
        });
      } else {
        const errorData = await response.json();
        throw new Error(errorData.error || '登录失败');
      }
    } catch (err) {
      alert("登录失败：" + (err?.message || "未知错误"));
    } finally {
      if (submitBtn) { submitBtn.disabled = false; submitBtn.textContent = "登录"; }
    }
  });

  // 使用 Supabase 注册（邮箱/密码）
  document.getElementById("registerForm")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    const account = document.getElementById("registerAccount").value.trim();
    const pwd = document.getElementById("registerPassword").value;
    if (!account || !pwd) { alert("请输入账号和密码"); return; }
    
    const submitBtn = document.querySelector("#registerForm button[type='submit']");
    if (submitBtn) { submitBtn.disabled = true; submitBtn.textContent = "注册中…"; }
    
    try {
      const response = await fetch('/api/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ account, password: pwd })
      });
      
      if (response.ok) {
        const data = await response.json();
        alert("注册成功，已自动登录");
        closeModal(registerModal);
        
        // 更新本地存储的用户信息
        window.authManager.setCurrentUser({
          id: data.user_id,
          nickname: data.nickname,
          email: account
        });
        
        // 触发登录事件，通知所有页面更新状态
        window.dispatchEvent(new CustomEvent('userLogin', { detail: data }));
        
        // 更新导航栏状态
        window.authManager.updateNavbarStatus(true, {
          id: data.user_id,
          nickname: data.nickname,
          email: account
        });
      } else {
        const errorData = await response.json();
        throw new Error(errorData.error || '注册失败');
      }
    } catch (err) {
      alert("注册失败：" + (err?.message || "未知错误"));
    } finally {
      if (submitBtn) { submitBtn.disabled = false; submitBtn.textContent = "注册"; }
    }
  });
}

// 歌曲详情页评论功能
(function () {
  // 检查是否在歌曲详情页
  const url = new URL(window.location.href);
  if (!url.pathname.endsWith("/song")) return;
  
  const songID = url.searchParams.get("id");
  if (!songID) return;
  
  const loginStatus = document.getElementById('loginStatus');
  const commentForm = document.getElementById('commentForm');
  const commentInput = document.getElementById('commentInput');
  const submitComment = document.getElementById('submitComment');
  const commentsList = document.getElementById('commentsList');
  const goToLogin = document.getElementById('goToLogin');
  
  // 检查用户登录状态
  async function checkAuthStatus() {
    try {
      const response = await fetch('/api/check_auth');
      if (!response.ok) {
        throw new Error('HTTP error: ' + response.status);
      }
      const data = await response.json();
      
      if (data.authenticated) {
        // 用户已登录，显示评论表单
        loginStatus.classList.add('hidden');
        commentForm.classList.remove('hidden');
      } else {
        // 用户未登录，显示登录提示
        loginStatus.classList.remove('hidden');
        commentForm.classList.add('hidden');
      }
    } catch (error) {
      console.error('检查登录状态失败:', error);
      // 默认显示登录提示
      loginStatus.classList.remove('hidden');
      commentForm.classList.add('hidden');
    }
  }
  
  // 加载评论
  async function loadComments() {
    try {
      const response = await fetch(`/api/comments?song_id=${songID}`);
      const comments = await response.json();
      
      if (Array.isArray(comments) && comments.length > 0) {
        renderComments(comments);
      } else {
        commentsList.innerHTML = '<div class="no-comments">暂无评论，快来发表第一条评论吧！</div>';
      }
    } catch (error) {
      console.error('加载评论失败:', error);
      commentsList.innerHTML = '<div class="no-comments">加载评论失败，请稍后重试</div>';
    }
  }
  
  // 渲染评论列表
  function renderComments(comments) {
    commentsList.innerHTML = '';
    
    comments.forEach(comment => {
      const commentElement = document.createElement('div');
      commentElement.className = 'comment-item';
      
      // 使用后端返回的用户昵称，如果没有昵称则使用默认值
      const userNickname = comment.nickname || comment.username || '用户' || '未知用户';
      
      // 安全处理日期
      let commentTime = '未知时间';
      try {
        if (comment.created_at) {
          commentTime = new Date(comment.created_at).toLocaleString('zh-CN');
        }
      } catch (error) {
        console.error('日期解析错误:', error);
        commentTime = comment.created_at || '未知时间';
      }
      
      // 获取评论内容
      const content = comment.content || '无内容';
      
      commentElement.innerHTML = `
        <div class="comment-header">
          <span class="comment-user">${userNickname}</span>
          <span class="comment-time">${commentTime}</span>
        </div>
        <div class="comment-content">${escapeHtml(content)}</div>
      `;
      
      commentsList.appendChild(commentElement);
    });
  }
  
  // HTML转义函数
  function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }
  
  // 提交评论
  async function handleSubmitComment() {
    const content = commentInput.value.trim();
    
    if (!content) {
      alert('请输入评论内容');
      return;
    }
    
    if (content.length > 500) {
      alert('评论内容不能超过500字');
      return;
    }
    
    submitComment.disabled = true;
    submitComment.textContent = '发表中...';
    
    try {
      const response = await fetch('/api/comments', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          song_id: parseInt(songID),
          content: content
        })
      });
      
      if (response.ok) {
        commentInput.value = '';
        await loadComments(); // 重新加载评论列表
        alert('评论发表成功！');
      } else {
        const errorData = await response.json();
        throw new Error(errorData.error || '发表评论失败');
      }
    } catch (error) {
      console.error('发表评论失败:', error);
      alert('发表评论失败：' + error.message);
    } finally {
      submitComment.disabled = false;
      submitComment.textContent = '发表评论';
    }
  }
  
  // 事件监听
  if (submitComment) {
    submitComment.addEventListener('click', handleSubmitComment);
  }
  
  if (commentInput) {
    commentInput.addEventListener('keydown', (e) => {
      if (e.ctrlKey && e.key === 'Enter') {
        handleSubmitComment();
      }
    });
  }
  
  if (goToLogin) {
    goToLogin.addEventListener('click', (e) => {
      e.preventDefault();
      // 直接打开登录模态框而不是跳转到首页
      if (window.openLoginModal) {
        window.openLoginModal();
      } else {
        // 如果模态框函数不存在，则跳转到首页
        window.location.href = '/';
      }
    });
  }
  
  // 初始化
  checkAuthStatus();
  loadComments();
});

// 立即初始化模态框交互（不等待DOMContentLoaded）
if (typeof initModalInteractions === 'function') {
  initModalInteractions();
}

// 初始化所有页面的登录状态
if (window.authManager && typeof window.authManager.initAllPagesAuth === 'function') {
  window.authManager.initAllPagesAuth();
// 确保模态框函数存在，如果不存在则重新初始化
if (!window.openLoginModal || !window.openRegisterModal) {
  console.warn('模态框函数未定义，重新初始化');
  if (typeof initModalInteractions === 'function') {
    initModalInteractions();
  }
}

