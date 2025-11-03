# 云端音乐存储系统部署指南

## 1. 系统架构概述

本系统实现了基于 Supabase Storage 的完整云端音乐存储方案：

```
用户前端 (HTML/JS/CSS)
    ↓
Go 后端服务器
    ↓
Supabase Storage (文件存储)
    ↓
Supabase Database (元数据存储)
    ↓
CDN 网络 (全球加速)
```

## 2. 部署前准备

### 2.1 环境要求

- **操作系统**: Windows/Linux/macOS
- **Go 版本**: 1.19+
- **Node.js**: 14+ (用于构建)
- **Supabase 账户**: 免费账户即可

### 2.2 文件结构检查

确保项目包含以下关键文件：

```
d:\AAAAA_PBL2\
├── main.go                    # 主程序入口
├── controller/                # 控制器层
│   ├── upload.go             # 上传控制器
│   └── music.go              # 音乐控制器
├── service/                   # 服务层
│   ├── upload.go             # 上传服务
│   └── music.go              # 音乐服务
├── db/                        # 数据库层
│   └── supabase.go           # Supabase 配置
├── web/                       # 前端文件
│   ├── index.html            # 首页
│   ├── upload.html           # 上传页面
│   ├── app.js                # 前端逻辑
│   └── styles.css            # 样式文件
├── go.mod                     # Go 模块配置
└── .env                       # 环境变量
```

## 3. Supabase 配置

### 3.1 创建 Supabase 项目

1. 访问 [Supabase](https://supabase.com)
2. 创建新项目
3. 记录项目 URL 和 API 密钥

### 3.2 配置存储桶

在 Supabase 控制台中：

1. 进入 "Storage" 页面
2. 创建名为 `music` 的存储桶
3. 配置权限为 "Public"
4. 设置文件大小限制为 100MB

### 3.3 配置数据库表

执行 `SUPABASE_TABLE_SETUP.md` 中的 SQL 语句创建所需表结构。

## 4. 环境配置

### 4.1 环境变量文件

创建 `.env` 文件并配置：

```env
# Supabase 配置
SUPABASE_URL=https://your-project-ref.supabase.co
SUPABASE_ANON_KEY=your-anon-key-here
SUPABASE_SERVICE_KEY=your-service-key-here

# 应用配置
PORT=8080
MUSIC_BUCKET=music
MAX_FILE_SIZE=100000000  # 100MB in bytes
```

### 4.2 验证环境配置

运行以下命令验证配置：

```bash
# 检查 Go 环境
go version

# 检查依赖
go mod tidy

# 构建应用
go build -o MusicPlayerWeb.exe
```

## 5. 部署步骤

### 5.1 本地部署（开发环境）

1. **启动服务器**:
   ```bash
   cd d:\AAAAA_PBL2
   .\MusicPlayerWeb.exe
   ```

2. **访问应用**:
   - 打开浏览器访问 `http://localhost:8080`
   - 测试上传功能：访问 `http://localhost:8080/upload`

### 5.2 生产环境部署

#### 选项1: 使用 Railway（推荐）

1. **连接 GitHub 仓库**
2. **配置环境变量**
3. **自动部署**

Railway 会自动检测 Go 项目并部署。

#### 选项2: 使用 Netlify

1. **构建静态文件**:
   ```bash
   # 构建前端资源
   # 确保 web/ 目录包含所有静态文件
   ```

2. **配置构建命令**:
   ```toml
   # netlify.toml
   [build]
     command = "go build -o MusicPlayerWeb.exe"
     publish = "."
   ```

#### 选项3: 手动服务器部署

1. **上传文件到服务器**
2. **安装 Go 运行环境**
3. **配置反向代理 (Nginx)**
4. **设置系统服务**

## 6. 功能测试

### 6.1 基础功能测试

1. **用户注册/登录**
   - 测试用户注册流程
   - 测试用户登录功能
   - 验证会话管理

2. **文件上传测试**
   - 上传各种音频格式
   - 测试文件大小限制
   - 验证元数据提取

3. **音乐播放测试**
   - 播放本地音乐
   - 播放云端音乐
   - 测试播放控制

### 6.2 高级功能测试

1. **权限控制测试**
   - 未登录用户访问限制
   - 用户只能访问自己的文件
   - 公开分享功能测试

2. **性能测试**
   - 多文件并发上传
   - 大文件上传稳定性
   - 播放流畅度测试

## 7. 监控和维护

### 7.1 日志监控

配置日志系统监控应用状态：

```go
// 日志配置示例
import "log"

func main() {
    // 设置日志输出
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    
    // 文件日志
    f, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    if err == nil {
        log.SetOutput(f)
        defer f.Close()
    }
}
```

### 7.2 性能监控

监控关键指标：
- 服务器响应时间
- 存储空间使用情况
- API 调用频率
- 错误率统计

## 8. 安全配置

### 8.1 网络安全

1. **HTTPS 配置**: 确保所有流量加密
2. **CORS 配置**: 限制跨域请求
3. **防火墙规则**: 限制不必要的端口访问

### 8.2 应用安全

1. **输入验证**: 所有用户输入验证
2. **文件安全**: 文件类型和大小验证
3. **会话安全**: 安全的会话管理

## 9. 备份和恢复

### 9.1 数据备份策略

1. **数据库备份**:
   - 每日自动备份
   - 保留最近7天备份
   - 异地备份存储

2. **文件备份**:
   - 重要文件版本控制
   - 定期快照备份
   - 跨地域冗余存储

### 9.2 恢复流程

定义清晰的灾难恢复流程：

1. **识别问题**
2. **恢复数据库**
3. **恢复文件**
4. **验证恢复**

## 10. 故障排除

### 10.1 常见问题

**问题1: 文件上传失败**
```bash
# 检查网络连接
ping supabase.co

# 检查 API 密钥
curl -H "Authorization: Bearer YOUR_ANON_KEY" \
     https://YOUR_PROJECT.supabase.co/storage/v1/bucket
```

**问题2: 播放器无法加载**
- 检查浏览器控制台错误
- 验证音频文件 URL 可访问
- 检查 CORS 配置

**问题3: 用户认证失败**
- 检查 Supabase 认证服务状态
- 验证会话管理逻辑
- 检查环境变量配置

### 10.2 调试工具

使用以下工具进行调试：

1. **浏览器开发者工具**: 检查网络请求和错误
2. **Postman**: 测试 API 接口
3. **Supabase Dashboard**: 监控数据库和存储

## 11. 扩展和优化

### 11.1 性能优化

1. **CDN 优化**: 配置更优的 CDN 策略
2. **缓存优化**: 实现多级缓存机制
3. **数据库优化**: 索引和查询优化

### 11.2 功能扩展

1. **移动端支持**: 响应式设计优化
2. **离线功能**: PWA 支持
3. **社交功能**: 分享和评论

## 12. 支持资源

### 12.1 文档链接

- [Supabase 文档](https://supabase.com/docs)
- [Go 语言文档](https://golang.org/doc/)
- [HTML5 Audio API](https://developer.mozilla.org/en-US/docs/Web/API/HTMLAudioElement)

### 12.2 社区支持

- [Supabase Discord](https://discord.supabase.com)
- [Go 社区论坛](https://forum.golangbridge.org/)
- [GitHub Issues](项目 GitHub 仓库)

这个部署指南提供了从开发到生产的完整流程，确保您的云端音乐存储系统顺利部署和运行。