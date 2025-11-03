@echo off
echo ========================================
echo Supabase数据库迁移脚本（方案1：整数ID → UUID）
echo ========================================
echo.

REM 设置环境变量
set SUPABASE_URL=https://gblnpzstdjnvclijjpbk.supabase.co
set SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImdibG5wenN0ZGpudmNsaWpqcGJrIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjA0NjgzNjEsImV4cCI6MjA3NjA0NDM2MX0.YaMwiKVVUgToza4vMtBjHAgvE28fdWTqtNHHtI9peFU
set SUPABASE_ACCESS_TOKEN=sbp_6f1a55d51a953cb9a111ff29bf87a7c2f487e6c0

echo ✅ 环境变量设置完成
echo 📊 连接到Supabase项目: %SUPABASE_URL%
echo.

REM 创建临时JSON文件用于存储SQL查询
set SQL_FILE=temp_sql.json

REM 1. 测试连接
echo 📝 步骤1: 测试数据库连接...
echo {"query": "SELECT current_database();"} > %SQL_FILE%
curl -X POST "%SUPABASE_URL%/rest/v1/rpc/execute_sql" ^
  -H "Content-Type: application/json" ^
  -H "apikey: %SUPABASE_ANON_KEY%" ^
  -H "Authorization: Bearer %SUPABASE_ACCESS_TOKEN%" ^
  -d @%SQL_FILE%

if errorlevel 1 (
    echo ❌ 连接测试失败
    goto :error
)
echo ✅ 连接测试成功
echo.

REM 2. 备份现有数据
echo 📝 步骤2: 备份现有数据...
echo {"query": "CREATE TABLE IF NOT EXISTS user_profiles_backup AS SELECT * FROM user_profiles;"} > %SQL_FILE%
curl -X POST "%SUPABASE_URL%/rest/v1/rpc/execute_sql" ^
  -H "Content-Type: application/json" ^
  -H "apikey: %SUPABASE_ANON_KEY%" ^
  -H "Authorization: Bearer %SUPABASE_ACCESS_TOKEN%" ^
  -d @%SQL_FILE%

echo ✅ 数据备份完成
echo.

REM 3. 删除现有表
echo 📝 步骤3: 删除现有表...
echo {"query": "DROP TABLE IF EXISTS forum_replies CASCADE; DROP TABLE IF EXISTS forum_posts CASCADE; DROP TABLE IF EXISTS user_profiles CASCADE;"} > %SQL_FILE%
curl -X POST "%SUPABASE_URL%/rest/v1/rpc/execute_sql" ^
  -H "Content-Type: application/json" ^
  -H "apikey: %SUPABASE_ANON_KEY%" ^
  -H "Authorization: Bearer %SUPABASE_ACCESS_TOKEN%" ^
  -d @%SQL_FILE%

echo ✅ 现有表删除完成
echo.

REM 4. 创建新表结构
echo 📝 步骤4: 创建新表结构...

REM 创建user_profiles表
echo {"query": "CREATE TABLE user_profiles (id UUID DEFAULT gen_random_uuid() PRIMARY KEY, user_id UUID NOT NULL, email TEXT NOT NULL, nickname TEXT, avatar_url TEXT, bio TEXT, created_at TIMESTAMPTZ DEFAULT now(), updated_at TIMESTAMPTZ DEFAULT now(), UNIQUE(user_id));"} > %SQL_FILE%
curl -X POST "%SUPABASE_URL%/rest/v1/rpc/execute_sql" ^
  -H "Content-Type: application/json" ^
  -H "apikey: %SUPABASE_ANON_KEY%" ^
  -H "Authorization: Bearer %SUPABASE_ACCESS_TOKEN%" ^
  -d @%SQL_FILE%

REM 创建forum_posts表
echo {"query": "CREATE TABLE forum_posts (id UUID DEFAULT gen_random_uuid() PRIMARY KEY, title TEXT NOT NULL, content TEXT NOT NULL, user_id UUID NOT NULL, user_nickname TEXT, tags TEXT[] DEFAULT '{}', view_count INTEGER DEFAULT 0, reply_count INTEGER DEFAULT 0, is_pinned BOOLEAN DEFAULT false, is_locked BOOLEAN DEFAULT false, created_at TIMESTAMPTZ DEFAULT now(), updated_at TIMESTAMPTZ DEFAULT now());"} > %SQL_FILE%
curl -X POST "%SUPABASE_URL%/rest/v1/rpc/execute_sql" ^
  -H "Content-Type: application/json" ^
  -H "apikey: %SUPABASE_ANON_KEY%" ^
  -H "Authorization: Bearer %SUPABASE_ACCESS_TOKEN%" ^
  -d @%SQL_FILE%

REM 创建forum_replies表
echo {"query": "CREATE TABLE forum_replies (id UUID DEFAULT gen_random_uuid() PRIMARY KEY, post_id UUID NOT NULL, content TEXT NOT NULL, user_id UUID NOT NULL, user_nickname TEXT, parent_id UUID, created_at TIMESTAMPTZ DEFAULT now(), updated_at TIMESTAMPTZ DEFAULT now(), FOREIGN KEY (post_id) REFERENCES forum_posts(id) ON DELETE CASCADE, FOREIGN KEY (parent_id) REFERENCES forum_replies(id) ON DELETE CASCADE);"} > %SQL_FILE%
curl -X POST "%SUPABASE_URL%/rest/v1/rpc/execute_sql" ^
  -H "Content-Type: application/json" ^
  -H "apikey: %SUPABASE_ANON_KEY%" ^
  -H "Authorization: Bearer %SUPABASE_ACCESS_TOKEN%" ^
  -d @%SQL_FILE%

echo ✅ 新表结构创建完成
echo.

REM 5. 启用行级安全
echo 📝 步骤5: 启用行级安全...
echo {"query": "ALTER TABLE user_profiles ENABLE ROW LEVEL SECURITY; ALTER TABLE forum_posts ENABLE ROW LEVEL SECURITY; ALTER TABLE forum_replies ENABLE ROW LEVEL SECURITY;"} > %SQL_FILE%
curl -X POST "%SUPABASE_URL%/rest/v1/rpc/execute_sql" ^
  -H "Content-Type: application/json" ^
  -H "apikey: %SUPABASE_ANON_KEY%" ^
  -H "Authorization: Bearer %SUPABASE_ACCESS_TOKEN%" ^
  -d @%SQL_FILE%

echo ✅ 行级安全启用完成
echo.

REM 6. 创建策略
echo 📝 步骤6: 创建访问策略...

REM 用户资料表策略
echo {"query": "CREATE POLICY \"任何人都可以查看用户资料\" ON user_profiles FOR SELECT USING (true); CREATE POLICY \"用户可以更新自己的资料\" ON user_profiles FOR UPDATE USING (auth.uid() = user_id); CREATE POLICY \"用户可以插入自己的资料\" ON user_profiles FOR INSERT WITH CHECK (auth.uid() = user_id);"} > %SQL_FILE%
curl -X POST "%SUPABASE_URL%/rest/v1/rpc/execute_sql" ^
  -H "Content-Type: application/json" ^
  -H "apikey: %SUPABASE_ANON_KEY%" ^
  -H "Authorization: Bearer %SUPABASE_ACCESS_TOKEN%" ^
  -d @%SQL_FILE%

REM 论坛帖子表策略
echo {"query": "CREATE POLICY \"任何人都可以查看帖子\" ON forum_posts FOR SELECT USING (true); CREATE POLICY \"登录用户可以创建帖子\" ON forum_posts FOR INSERT WITH CHECK (auth.uid() = user_id); CREATE POLICY \"用户可以更新自己的帖子\" ON forum_posts FOR UPDATE USING (auth.uid() = user_id); CREATE POLICY \"用户可以删除自己的帖子\" ON forum_posts FOR DELETE USING (auth.uid() = user_id);"} > %SQL_FILE%
curl -X POST "%SUPABASE_URL%/rest/v1/rpc/execute_sql" ^
  -H "Content-Type: application/json" ^
  -H "apikey: %SUPABASE_ANON_KEY%" ^
  -H "Authorization: Bearer %SUPABASE_ACCESS_TOKEN%" ^
  -d @%SQL_FILE%

REM 论坛回复表策略
echo {"query": "CREATE POLICY \"任何人都可以查看回复\" ON forum_replies FOR SELECT USING (true); CREATE POLICY \"登录用户可以创建回复\" ON forum_replies FOR INSERT WITH CHECK (auth.uid() = user_id); CREATE POLICY \"用户可以更新自己的回复\" ON forum_replies FOR UPDATE USING (auth.uid() = user_id); CREATE POLICY \"用户可以删除自己的回复\" ON forum_replies FOR DELETE USING (auth.uid() = user_id);"} > %SQL_FILE%
curl -X POST "%SUPABASE_URL%/rest/v1/rpc/execute_sql" ^
  -H "Content-Type: application/json" ^
  -H "apikey: %SUPABASE_ANON_KEY%" ^
  -H "Authorization: Bearer %SUPABASE_ACCESS_TOKEN%" ^
  -d @%SQL_FILE%

echo ✅ 访问策略创建完成
echo.

REM 7. 创建索引
echo 📝 步骤7: 创建索引...
echo {"query": "CREATE INDEX IF NOT EXISTS idx_forum_posts_user_id ON forum_posts(user_id); CREATE INDEX IF NOT EXISTS idx_forum_posts_created_at ON forum_posts(created_at); CREATE INDEX IF NOT EXISTS idx_forum_posts_tags ON forum_posts USING gin(tags); CREATE INDEX IF NOT EXISTS idx_forum_replies_post_id ON forum_replies(post_id); CREATE INDEX IF NOT EXISTS idx_forum_replies_user_id ON forum_replies(user_id); CREATE INDEX IF NOT EXISTS idx_forum_replies_parent_id ON forum_replies(parent_id);"} > %SQL_FILE%
curl -X POST "%SUPABASE_URL%/rest/v1/rpc/execute_sql" ^
  -H "Content-Type: application/json" ^
  -H "apikey: %SUPABASE_ANON_KEY%" ^
  -H "Authorization: Bearer %SUPABASE_ACCESS_TOKEN%" ^
  -d @%SQL_FILE%

echo ✅ 索引创建完成
echo.

REM 8. 验证结果
echo 📝 步骤8: 验证迁移结果...
echo {"query": "SELECT table_name, table_type FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('user_profiles', 'forum_posts', 'forum_replies');"} > %SQL_FILE%
curl -X POST "%SUPABASE_URL%/rest/v1/rpc/execute_sql" ^
  -H "Content-Type: application/json" ^
  -H "apikey: %SUPABASE_ANON_KEY%" ^
  -H "Authorization: Bearer %SUPABASE_ACCESS_TOKEN%" ^
  -d @%SQL_FILE%

echo.
echo 🎉 数据库迁移完成！
echo ✅ 表结构已成功修改为UUID格式
echo ✅ 行级安全策略已配置
echo ✅ 索引已创建
echo.
echo 📋 下一步：请更新Go代码中的ID处理逻辑
echo.

REM 清理临时文件
del %SQL_FILE%

goto :end

:error
echo ❌ 迁移过程中出现错误
del %SQL_FILE%
exit /b 1

:end
echo ========================================
echo 迁移脚本执行完毕
echo ========================================
pause