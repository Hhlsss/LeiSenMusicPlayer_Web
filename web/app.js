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
    
    // 监听用户信息更新事件
    window.addEventListener('userProfileUpdated', (event) => {
      this.updateNavbarStatus(true, event.detail);
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

      // 用歌曲封面替换首页轮播图
      const slidesWrap = document.querySelector(".slides");
      if (slidesWrap) {
        const covers = list.filter(item => item.hasCover);
        if (covers.length) {
          slidesWrap.innerHTML = "";
          // 随机抽样 5 个封面
          const shuffled = covers.slice().sort(() => Math.random() - 0.5);
          shuffled.slice(0, 5).forEach((item, idx) => {
            const div = document.createElement("div");
            div.className = "slide" + (idx === 0 ? " active" : "");
            div.style.backgroundImage = "url('/api/cover?id=" + item.id + "')";
            slidesWrap.appendChild(div);
          });
          // 重新初始化轮播
          if (window.initCarousel) window.initCarousel();
        }
      }
    })
    .catch(() => {});
})();

// 歌曲页：根据 id 加载音频、封面与歌词
(function () {
  const url = new URL(window.location.href);
  const id = url.pathname.endsWith("/song") ? url.searchParams.get("id") : null;
  const audio = document.getElementById("audio");
  if (!id || !audio) return;

  audio.src = "/api/audio?id=" + encodeURIComponent(id);

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

  const lyricsEl = document.querySelector(".lyrics");
  if (lyricsEl) {
    fetch("/api/lyrics_raw?id=" + encodeURIComponent(id))
      .then(res => res.json())
      .then(j => {
        if (!j || !j.lyrics) { lyricsEl.textContent = "暂无歌词"; return; }
        const lines = [];
        const re = /^\[(\d{1,2}):(\d{2})(?:\.(\d{1,3}))?\](.*)$/;
        j.lyrics.split(/\r?\n/).forEach(raw => {
          const m = raw.match(re);
          if (m) {
            const min = parseInt(m[1], 10);
            const sec = parseInt(m[2], 10);
            const ms = m[3] ? parseInt(m[3], 10) : 0;
            const t = min * 60 + sec + ms / 1000;
            const text = m[4].trim();
            if (text) lines.push({ t, text });
          }
        });
        lines.sort((a, b) => a.t - b.t);
        if (!lines.length) { lyricsEl.textContent = "暂无歌词"; return; }
        // 渲染为行
        lyricsEl.innerHTML = "";
        lines.forEach((ln, i) => {
          const p = document.createElement("p");
          p.textContent = ln.text;
          p.dataset.t = ln.t;
          lyricsEl.appendChild(p);
        });
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
        
        // 点击歌词跳转播放
        lyricsEl.addEventListener('click', (e) => {
          if (e.target.tagName === 'P' && !e.target.classList.contains('active')) {
            const targetTime = parseFloat(e.target.dataset.t);
            if (!isNaN(targetTime)) {
              audio.currentTime = targetTime;
              if (audio.paused) {
                audio.play();
              }
            }
          }
        });
        
        const highlight = (cur) => {
          let idx = 0;
          for (let i = 0; i < lines.length; i++) {
            if (cur >= lines[i].t) idx = i; else break;
          }
          const children = Array.from(lyricsEl.children);
          children.forEach((el, i) => el.classList.toggle("active", i === idx));
          
          // 如果用户没有在滚动，自动滚动到高亮歌词
          if (!isUserScrolling && children[idx]) {
            const activeElement = children[idx];
            const containerHeight = lyricsEl.clientHeight;
            const elementHeight = activeElement.offsetHeight;
            const elementTop = activeElement.offsetTop;
            
            // 计算目标滚动位置，使高亮歌词显示在中央
            const targetScrollTop = elementTop - (containerHeight / 2) - (elementHeight * 6);
            
            // 确保滚动位置在合理范围内
            const maxScrollTop = lyricsEl.scrollHeight - containerHeight;
            const adjustedScrollTop = Math.max(0, Math.min(targetScrollTop, maxScrollTop));
            
            // 平滑滚动到目标位置
            lyricsEl.scrollTo({
              top: adjustedScrollTop,
              behavior: 'smooth'
            });
          }
        };
        audio.addEventListener("timeupdate", () => highlight(audio.currentTime));
        // 初始高亮
        highlight(0);
      })
      .catch(() => { lyricsEl.textContent = "暂无歌词"; });
  }
})();

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
(function () {
  const overlay = document.getElementById("modalOverlay");
  const loginModal = document.getElementById("loginModal");
  const registerModal = document.getElementById("registerModal");

  const openModal = (modal) => {
    overlay.classList.remove("hidden");
    modal.classList.remove("hidden");
  };
  const closeModal = (modal) => {
    overlay.classList.add("hidden");
    modal.classList.add("hidden");
  };

  document.getElementById("loginOpen")?.addEventListener("click", () => openModal(loginModal));
  document.getElementById("registerOpen")?.addEventListener("click", () => openModal(registerModal));

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
})();

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
      // 跳转到首页进行登录
      window.location.href = '/';
    });
  }
  
  // 初始化
  checkAuthStatus();
  loadComments();
})();