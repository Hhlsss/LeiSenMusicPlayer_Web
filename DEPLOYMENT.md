# 音乐播放器项目部署指南

## 🚀 部署方案

### 方案1：Netlify（前端）+ Railway（后端） - 推荐

#### 后端部署到 Railway
1. **注册 Railway 账户** (https://railway.app)
2. **连接 GitHub 仓库**
3. **创建新项目**并选择此仓库
4. **自动部署** - Railway 会自动检测 Go 项目并部署
5. **获取后端 URL** - 部署完成后记下 Railway 提供的域名

#### 前端部署到 Netlify
1. **注册 Netlify 账户** (https://netlify.com)
2. **连接 GitHub 仓库**
3. **配置构建设置**：
   - 构建命令: `echo "Building static site"`
   - 发布目录: `web`
4. **环境变量**：
   ```
   BACKEND_URL=https://your-railway-service.up.railway.app
   ```
5. **更新重定向规则**：编辑 `netlify.toml` 中的后端 URL

### 方案2：全栈部署到 Railway

1. **部署到 Railway**（同上）
2. **配置构建**：Railway 会自动处理 Go 后端
3. **访问应用**：使用 Railway 提供的域名

## 📁 项目结构

```
MusicPlayerWeb/
├── main.go                 # Go 后端入口
├── go.mod                 # Go 依赖管理
├── web/                   # 前端静态文件
│   ├── index.html
│   ├── app.js
│   ├── styles.css
│   └── ...
├── controller/            # 后端控制器
├── service/              # 业务逻辑
├── db/                   # 数据库操作
├── netlify.toml          # Netlify 配置
└── railway.toml          # Railway 配置
```

## 🔧 环境变量配置

### Railway 环境变量
```bash
PORT=8080
MUSIC_DIR=./music          # 音乐文件目录
```

### Netlify 环境变量
```bash
BACKEND_URL=你的后端服务地址
```

## 🎯 部署步骤详解

### 后端部署（Railway）
1. Fork 或上传代码到 GitHub
2. 访问 https://railway.app
3. 点击 "New Project" → "Deploy from GitHub repo"
4. 选择你的仓库
5. Railway 会自动构建和部署
6. 记下部署后的域名（如：https://your-app.up.railway.app）

### 前端部署（Netlify）
1. 访问 https://app.netlify.com
2. 点击 "Add new site" → "Import an existing project"
3. 连接 GitHub 账户并选择仓库
4. 配置构建设置：
   - Base directory: (留空)
   - Build command: `echo "Static site"`
   - Publish directory: `web`
5. 点击 "Deploy site"
6. 在 Site settings → Environment variables 中添加 BACKEND_URL

## 🌐 访问应用

部署完成后：
- **Netlify 前端**: https://your-site.netlify.app
- **Railway 后端**: https://your-app.up.railway.app

## 🔄 更新部署

代码推送到 GitHub 后，Railway 和 Netlify 会自动重新部署。

## ❗ 注意事项

1. **音乐文件处理**：后端需要访问音乐文件，确保音乐文件在部署环境中可用
2. **CORS 配置**：前后端分离需要配置 CORS
3. **文件路径**：部署环境中的文件路径可能与本地不同
4. **环境变量**：敏感信息请使用环境变量，不要硬编码在代码中

## 📞 故障排除

### 常见问题
1. **构建失败**：检查 Go 版本兼容性
2. **API 404**：确认后端服务 URL 正确
3. **静态资源 404**：检查 Netlify 发布目录设置
4. **CORS 错误**：配置后端 CORS 中间件

### 日志查看
- Railway: 项目页面 → "Deployments" → 选择部署 → "Logs"
- Netlify: Site settings → "Deploys" → 选择部署 → "Deploy log"