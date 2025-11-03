# Supabase Storage 配置指南

## 1. Supabase Storage 权限配置

### 1.1 创建存储桶

在 Supabase 控制台中创建名为 `music` 的存储桶：

1. 登录 Supabase 控制台
2. 选择您的项目
3. 进入 "Storage" 页面
4. 点击 "New Bucket"
5. 填写以下信息：
   - **Bucket Name**: `music`
   - **Public**: ✅ 勾选（允许公开读取）
   - **File Size Limit**: 100MB
   - **Allowed MIME Types**: 
     - `audio/*` (所有音频文件)
     - `application/octet-stream` (二进制文件)

### 1.2 配置 Row Level Security (RLS)

在 SQL 编辑器中执行以下 SQL 语句来配置权限：

```sql
-- 启用存储桶的 RLS
ALTER TABLE storage.objects ENABLE ROW LEVEL SECURITY;

-- 允许公开读取音乐文件
CREATE POLICY "允许公开读取音乐文件" ON storage.objects
FOR SELECT USING (
  bucket_id = 'music' AND 
  (storage.foldername(name))[1] = 'public'
);

-- 允许认证用户上传文件
CREATE POLICY "允许认证用户上传音乐文件" ON storage.objects
FOR INSERT WITH CHECK (
  bucket_id = 'music' AND 
  auth.role() = 'authenticated'
);

-- 允许用户删除自己的文件
CREATE POLICY "允许用户删除自己的音乐文件" ON storage.objects
FOR DELETE USING (
  bucket_id = 'music' AND 
  auth.uid()::text = (storage.foldername(name))[1]
);
```

### 1.3 配置存储桶权限

在存储桶设置中配置以下权限：

- **Public**: 允许公开读取
- **Authenticated**: 允许认证用户上传和删除
- **Service Role**: 允许服务角色管理所有文件

## 2. 数据库表配置

### 2.1 创建音乐文件表

```sql
-- 创建音乐文件表
CREATE TABLE IF NOT EXISTS music_files (
    id VARCHAR PRIMARY KEY,
    title VARCHAR NOT NULL,
    artist VARCHAR NOT NULL,
    album VARCHAR,
    file_name VARCHAR NOT NULL,
    file_size BIGINT NOT NULL,
    file_type VARCHAR NOT NULL,
    storage_path VARCHAR NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_id UUID NOT NULL REFERENCES auth.users(id),
    is_public BOOLEAN DEFAULT false
);

-- 启用 RLS
ALTER TABLE music_files ENABLE ROW LEVEL SECURITY;

-- 用户只能访问自己的文件
CREATE POLICY "用户只能访问自己的音乐文件" ON music_files
FOR ALL USING (auth.uid() = user_id);

-- 允许公开读取公开文件
CREATE POLICY "允许公开读取公开音乐文件" ON music_files
FOR SELECT USING (is_public = true);
```

## 3. 环境变量配置

在项目根目录的 `.env` 文件中配置以下环境变量：

```env
# Supabase 配置
SUPABASE_URL=https://your-project-ref.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_KEY=your-service-key

# 应用配置
MUSIC_BUCKET=music
MAX_FILE_SIZE=100MB
ALLOWED_FILE_TYPES=audio/mpeg,audio/flac,audio/wav,audio/ogg,audio/x-m4a,audio/aac,audio/x-ms-wma,audio/webm,application/octet-stream
```

## 4. 安全配置

### 4.1 文件上传安全

1. **文件类型验证**: 只允许音频文件类型
2. **文件大小限制**: 最大 100MB
3. **病毒扫描**: 集成 ClamAV 或第三方扫描服务
4. **频率限制**: 防止恶意上传攻击

### 4.2 访问控制

1. **用户认证**: 只有登录用户才能上传
2. **权限验证**: 用户只能访问自己的文件
3. **公开分享**: 支持文件公开分享功能
4. **API 限流**: 防止 API 滥用

## 5. 性能优化

### 5.1 CDN 配置

Supabase Storage 自动集成 CDN，可以通过以下方式优化：

```go
// 获取优化的 CDN URL
func GetOptimizedMusicURL(storagePath string) string {
    baseURL := os.Getenv("SUPABASE_URL")
    // 使用 CDN 域名替换原始域名
    return strings.Replace(baseURL, "supabase.co", "cdn.supabase.co", 1) + 
           "/storage/v1/object/public/music/" + storagePath
}
```

### 5.2 缓存策略

1. **浏览器缓存**: 设置合适的 Cache-Control 头
2. **CDN 缓存**: 配置缓存规则
3. **数据库缓存**: 热门音乐信息缓存

## 6. 监控和日志

### 6.1 监控指标

- 上传成功率
- 平均上传时间
- 存储空间使用情况
- 播放失败率
- API 调用频率

### 6.2 错误处理

```go
func HandleUploadError(err error) string {
    switch {
    case errors.Is(err, ErrFileTooLarge):
        return "文件大小超过限制"
    case errors.Is(err, ErrInvalidFormat):
        return "不支持的音频格式"
    case errors.Is(err, ErrUploadFailed):
        return "上传失败，请重试"
    default:
        return "系统错误，请联系管理员"
    }
}
```

## 7. 备份和恢复

### 7.1 数据备份

1. **定期备份**: 每天自动备份数据库
2. **存储桶版本控制**: 启用文件版本控制
3. **跨地域备份**: 重要文件多地域备份

### 7.2 灾难恢复

1. **恢复流程**: 定义清晰的恢复流程
2. **测试恢复**: 定期测试恢复流程
3. **文档记录**: 记录所有配置和流程

## 8. 部署检查清单

### 8.1 部署前检查

- [ ] Supabase 项目已创建
- [ ] 存储桶已配置
- [ ] RLS 策略已启用
- [ ] 数据库表已创建
- [ ] 环境变量已配置
- [ ] 文件上传功能已测试
- [ ] 权限控制已测试

### 8.2 部署后验证

- [ ] 文件上传功能正常
- [ ] 文件播放功能正常
- [ ] 权限控制正常
- [ ] 错误处理正常
- [ ] 性能测试通过

## 9. 故障排除

### 9.1 常见问题

**问题1**: 文件上传失败
- **原因**: 存储桶权限配置错误
- **解决**: 检查存储桶 RLS 策略

**问题2**: 文件无法播放
- **原因**: CDN 缓存问题
- **解决**: 清除 CDN 缓存或等待缓存过期

**问题3**: 权限验证失败
- **原因**: 用户认证问题
- **解决**: 检查用户会话和认证状态

### 9.2 日志分析

查看 Supabase 日志和应用程序日志来诊断问题：

```bash
# 查看应用程序日志
tail -f /var/log/application.log

# 查看 Supabase 日志
# 在 Supabase 控制台的 Logs 页面查看
```

## 10. 扩展功能

### 10.1 未来扩展

1. **智能推荐**: 基于播放历史推荐音乐
2. **歌单同步**: 多设备歌单同步
3. **离线下载**: 支持离线播放
4. **音质选择**: 多音质版本支持

### 10.2 第三方集成

1. **支付系统**: 付费音乐支持
2. **社交分享**: 音乐分享功能
3. **版权管理**: 版权验证系统

这个配置指南提供了完整的 Supabase Storage 集成方案，确保您的音乐存储系统安全、高效地运行。