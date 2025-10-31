# 本地音乐上传到Supabase功能说明

## 功能概述

本项目已成功实现将本地音乐文件上传到Supabase数据库的完整功能，包括：

1. **前端上传界面** - 现代化的拖放上传界面
2. **后端API处理** - 完整的文件上传、元数据提取、存储管理
3. **Supabase集成** - 文件存储到Supabase Storage，元数据保存到数据库
4. **用户认证** - 只有登录用户才能上传音乐

## 功能特性

### 前端功能
- ✅ 拖放文件上传支持
- ✅ 多文件批量上传
- ✅ 实时上传进度显示
- ✅ 文件类型验证（音频文件）
- ✅ 文件大小限制（100MB）
- ✅ 上传统计信息
- ✅ 响应式设计

### 后端功能
- ✅ 文件上传到Supabase Storage
- ✅ 音乐元数据自动提取（标题、艺术家、专辑）
- ✅ 数据库记录管理
- ✅ 用户权限验证
- ✅ 文件删除功能
- ✅ 上传状态查询

## 技术实现

### 文件结构
```
controller/upload.go      # 上传控制器
service/upload.go         # 上传业务逻辑
web/upload.html          # 上传页面
main.go                  # 路由配置
```

### API接口

#### 1. 上传页面
- **URL**: `/upload`
- **方法**: GET
- **功能**: 显示音乐上传界面

#### 2. 音乐文件上传
- **URL**: `/api/upload/music`
- **方法**: POST
- **参数**: multipart/form-data (file字段)
- **功能**: 上传音乐文件到Supabase

#### 3. 上传状态查询
- **URL**: `/api/upload/status`
- **方法**: GET
- **功能**: 获取用户上传统计信息

#### 4. 获取用户音乐文件
- **URL**: `/api/upload/music`
- **方法**: GET
- **功能**: 获取用户上传的音乐文件列表

#### 5. 删除音乐文件
- **URL**: `/api/upload/music/{id}`
- **方法**: DELETE
- **功能**: 删除指定的音乐文件

#### 6. 播放上传的音乐
- **URL**: `/api/upload/play`
- **方法**: GET
- **参数**: id (音乐文件ID)
- **功能**: 播放上传的音乐文件

## 使用说明

### 1. 启动服务器
```bash
cd d:\AAAAA_PBL2
.\MusicPlayerWeb.exe
```

### 2. 访问上传页面
打开浏览器访问：`http://localhost:8080/upload`

### 3. 登录账户
- 需要先登录才能上传音乐
- 如果未登录，系统会提示登录

### 4. 上传音乐
1. 点击"选择文件"或直接拖放音乐文件到上传区域
2. 支持MP3、FLAC、WAV、OGG、M4A、AAC、WMA、WebM格式
3. 单个文件最大100MB
4. 点击"开始上传"按钮开始上传

### 5. 查看上传结果
- 上传完成后可以在文件列表中查看状态
- 统计信息区域显示总文件数、已上传数、失败数

## 数据库设计

### music_files表结构
```sql
CREATE TABLE music_files (
    id VARCHAR PRIMARY KEY,
    title VARCHAR NOT NULL,
    artist VARCHAR NOT NULL,
    album VARCHAR,
    file_name VARCHAR NOT NULL,
    file_size BIGINT NOT NULL,
    file_type VARCHAR NOT NULL,
    storage_path VARCHAR NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES auth.users(id)
);
```

## Supabase配置

### 环境变量
确保以下环境变量已正确配置：
```
SUPABASE_URL=你的Supabase项目URL
SUPABASE_ANON_KEY=你的Supabase匿名密钥
```

### Storage配置
需要在Supabase中创建名为`music`的存储桶，并设置适当的权限。

## 技术细节

### 元数据提取
- 使用`github.com/dhowden/tag`库提取音乐文件元数据
- 支持ID3标签、FLAC元数据等
- 自动填充标题、艺术家、专辑信息

### 文件存储
- 文件存储在Supabase Storage的`music`桶中
- 存储路径格式：`music/{user_id}/{timestamp}_{filename}`
- 支持公开访问和权限控制

### 错误处理
- 完整的错误处理和用户反馈
- 文件类型验证
- 文件大小限制
- 网络异常处理

## 安全考虑

1. **用户认证** - 只有登录用户才能上传
2. **文件类型验证** - 只允许音频文件
3. **文件大小限制** - 防止大文件攻击
4. **路径安全** - 防止路径遍历攻击
5. **数据库权限** - 使用行级安全策略

## 扩展功能建议

### 未来可添加的功能
1. **音乐播放列表** - 基于上传音乐创建播放列表
2. **音乐分享** - 分享上传的音乐给其他用户
3. **音乐搜索** - 在用户上传的音乐中搜索
4. **批量操作** - 批量删除、下载音乐文件
5. **音乐转码** - 自动转码为统一格式

## 故障排除

### 常见问题

1. **上传失败**
   - 检查网络连接
   - 确认Supabase配置正确
   - 检查文件大小和类型

2. **无法访问上传页面**
   - 确认服务器已启动
   - 检查端口8080是否被占用

3. **登录问题**
   - 确认用户已注册并登录
   - 检查浏览器Cookie设置

### 日志查看
服务器日志会显示详细的错误信息，可用于故障诊断。

## 总结

本地音乐上传到Supabase的功能已完整实现，提供了用户友好的界面和强大的后端支持。该功能将本地音乐管理扩展到了云端，为用户提供了更好的音乐存储和访问体验。