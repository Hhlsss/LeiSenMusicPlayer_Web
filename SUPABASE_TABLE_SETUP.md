# Supabase 表设置指南

## 问题描述
您的音乐播放器应用需要 `user_profiles` 表来存储用户昵称信息，但该表在 Supabase 中不存在。

## 解决方案

### 方法一：通过 Supabase 仪表板创建表（推荐）

1. **登录 Supabase 控制台**
   - 访问：https://supabase.com/dashboard
   - 使用您的 GitHub 或 Google 账号登录

2. **选择项目**
   - 选择项目：`gblnpzstdjnvclijjpbk`
   - 项目 URL：https://gblnpzstdjnvclijjpbk.supabase.co

3. **进入 Table Editor**
   - 在左侧菜单中点击 "Table Editor"
   - 点击 "Create a new table"

4. **创建表结构**
   - **表名**: `user_profiles`
   - **字段设置**:
     
     | 字段名 | 类型 | 约束 | 默认值 |
     |--------|------|------|--------|
     | user_id | integer | Primary Key | - |
     | email | text | Not Null | - |
     | nickname | text | Not Null | - |
     | created_at | timestamptz | - | now() |
     | updated_at | timestamptz | - | now() |

5. **保存表**
   - 点击 "Save" 按钮创建表

### 方法二：通过 SQL Editor 创建表

1. **进入 SQL Editor**
   - 在左侧菜单中点击 "SQL Editor"
   - 点击 "New query"

2. **执行 SQL 语句**
   ```sql
   CREATE TABLE user_profiles (
       user_id INTEGER PRIMARY KEY,
       email VARCHAR NOT NULL,
       nickname VARCHAR NOT NULL,
       created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
       updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
   );
   ```

3. **运行查询**
   - 点击 "Run" 或按 Ctrl+Enter 执行

### 方法三：启用 RLS 和设置权限（可选但推荐）

创建表后，建议设置行级安全策略：

```sql
-- 启用行级安全
ALTER TABLE user_profiles ENABLE ROW LEVEL SECURITY;

-- 创建策略：用户只能访问自己的数据
CREATE POLICY "用户只能访问自己的个人资料" ON user_profiles
    FOR ALL USING (auth.uid() = user_id);
```

## 验证表是否创建成功

1. **在 Table Editor 中查看**
   - 表应该出现在 Table Editor 的列表中

2. **通过 API 测试**
   ```bash
   curl "https://gblnpzstdjnvclijjpbk.supabase.co/rest/v1/user_profiles?limit=1" \
     -H "apikey: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImdibG5wenN0ZGpudmNsaWpqcGJrIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjA0NjgzNjEsImV4cCI6MjA3NjA0NDM2MX0.YaMwiKVVUgToza4vMtBjHAgvE28fdWTqtNHHtI9peFU" \
     -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImdibG5wenN0ZGpudmNsaWpqcGJrIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjA0NjgzNjEsImV4cCI6MjA3NjA0NDM2MX0.YaMwiKVVUgToza4vMtBjHAgvE28fdWTqtNHHtI9peFU"
   ```

## 故障排除

### 常见问题

1. **"表不存在"错误**
   - 确保表名拼写正确：`user_profiles`
   - 检查表是否在正确的项目中创建

2. **权限错误**
   - 确保使用了正确的 API 密钥
   - 检查 RLS 策略是否过于严格

3. **连接错误**
   - 检查网络连接
   - 验证 Supabase URL 和 API 密钥是否正确

### 联系支持
如果遇到问题，可以通过以下方式获取帮助：
- Supabase 官方文档：https://supabase.com/docs
- GitHub Issues：在项目仓库中创建 issue
- Discord 社区：加入 Supabase Discord 社区

## 下一步

创建表后，您的昵称修改功能应该可以正常工作了。如果仍有问题，请检查：

1. 应用代码中的错误处理
2. 网络连接和 API 调用
3. 用户认证状态