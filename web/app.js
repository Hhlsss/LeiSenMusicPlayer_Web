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

  // 播放时封面旋转
  const discCover = document.querySelector(".disc-cover");
  if (discCover) {
    audio.addEventListener("play",   () => discCover.classList.add("rotating"));
    audio.addEventListener("pause",  () => discCover.classList.remove("rotating"));
    audio.addEventListener("ended",  () => discCover.classList.remove("rotating"));
  }

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

  audio.addEventListener("timeupdate", () => {
    curTime.textContent = fmt(audio.currentTime);
    if (!isSeeking && isFinite(audio.duration)) {
      progress.value = Math.floor(audio.currentTime);
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
  });

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
    if (!window.supabase) return; // 非首页等无需 Supabase 的页面可忽略
    const url = window.__SUPABASE_URL__;
    const key = window.__SUPABASE_ANON_KEY__;
    if (!url || !key || url === 'YOUR_SUPABASE_URL' || key === 'YOUR_SUPABASE_ANON_KEY') {
      console.warn('Supabase 未配置，跳过初始化');
      return;
    }
    const client = window.supabase.createClient(url, key);
    window.supabaseClient = client;

    const loginBtn = document.getElementById('loginOpen');
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

  // 使用 Supabase 账号登录（邮箱/密码）
  document.getElementById("loginForm")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    const account = document.getElementById("loginAccount").value.trim();
    const pwd = document.getElementById("loginPassword").value;
    if (!account || !pwd) { alert("请输入账号和密码"); return; }

    const submitBtn = document.querySelector("#loginForm button[type='submit']");
    if (submitBtn) { submitBtn.disabled = true; submitBtn.textContent = "登录中…"; }

    try {
      if (!window.supabaseClient) throw new Error('Supabase 未初始化');
      const { data, error } = await window.supabaseClient.auth.signInWithPassword({ email: account, password: pwd });
      if (error) throw error;
      alert("登录成功");
      closeModal(loginModal);
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
    try {
      if (!window.supabaseClient) throw new Error('Supabase 未初始化');
      const { data, error } = await window.supabaseClient.auth.signUp({ email: account, password: pwd });
      if (error) throw error;
      alert("注册成功，请前往邮箱完成验证（如开启邮箱确认）");
      closeModal(registerModal);
    } catch (err) {
      alert("注册失败：" + (err?.message || "未知错误"));
    }
  });
})();