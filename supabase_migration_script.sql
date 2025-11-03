-- Supabase数据库表结构迁移脚本（方案1：整数ID → UUID）
-- 执行前请确保已备份数据

-- 1. 首先检查现有表结构
SELECT table_name, column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_schema = 'public' 
AND table_name IN ('user_profiles', 'forum_posts', 'forum_replies')
ORDER BY table_name, ordinal_position;

-- 2. 备份现有user_profiles表数据
CREATE TABLE IF NOT EXISTS user_profiles_backup AS SELECT * FROM user_profiles;

-- 3. 检查是否有forum_posts和forum_replies表，如果有则备份
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'forum_posts') THEN
        CREATE TABLE IF NOT EXISTS forum_posts_backup AS SELECT * FROM forum_posts;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'forum_replies') THEN
        CREATE TABLE IF NOT EXISTS forum_replies_backup AS SELECT * FROM forum_replies;
    END IF;
END$$;

-- 4. 删除现有表（如果存在）
DROP TABLE IF EXISTS forum_replies CASCADE;
DROP TABLE IF EXISTS forum_posts CASCADE;
DROP TABLE IF EXISTS user_profiles CASCADE;

-- 5. 重新创建user_profiles表（使用UUID）
CREATE TABLE user_profiles (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL,
    email TEXT NOT NULL,
    nickname TEXT,
    avatar_url TEXT,
    bio TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(user_id)
);

-- 6. 创建forum_posts表
CREATE TABLE forum_posts (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    user_id UUID NOT NULL,
    user_nickname TEXT,
    tags TEXT[] DEFAULT '{}',
    view_count INTEGER DEFAULT 0,
    reply_count INTEGER DEFAULT 0,
    is_pinned BOOLEAN DEFAULT false,
    is_locked BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 7. 创建forum_replies表
CREATE TABLE forum_replies (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    post_id UUID NOT NULL,
    content TEXT NOT NULL,
    user_id UUID NOT NULL,
    user_nickname TEXT,
    parent_id UUID,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    FOREIGN KEY (post_id) REFERENCES forum_posts(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES forum_replies(id) ON DELETE CASCADE
);

-- 8. 启用行级安全
ALTER TABLE user_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE forum_posts ENABLE ROW LEVEL SECURITY;
ALTER TABLE forum_replies ENABLE ROW LEVEL SECURITY;

-- 9. 创建用户资料表策略
CREATE POLICY "任何人都可以查看用户资料" ON user_profiles FOR SELECT USING (true);
CREATE POLICY "用户可以更新自己的资料" ON user_profiles FOR UPDATE USING (auth.uid() = user_id);
CREATE POLICY "用户可以插入自己的资料" ON user_profiles FOR INSERT WITH CHECK (auth.uid() = user_id);

-- 10. 创建论坛帖子表策略
CREATE POLICY "任何人都可以查看帖子" ON forum_posts FOR SELECT USING (true);
CREATE POLICY "登录用户可以创建帖子" ON forum_posts FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "用户可以更新自己的帖子" ON forum_posts FOR UPDATE USING (auth.uid() = user_id);
CREATE POLICY "用户可以删除自己的帖子" ON forum_posts FOR DELETE USING (auth.uid() = user_id);

-- 11. 创建论坛回复表策略
CREATE POLICY "任何人都可以查看回复" ON forum_replies FOR SELECT USING (true);
CREATE POLICY "登录用户可以创建回复" ON forum_replies FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "用户可以更新自己的回复" ON forum_replies FOR UPDATE USING (auth.uid() = user_id);
CREATE POLICY "用户可以删除自己的回复" ON forum_replies FOR DELETE USING (auth.uid() = user_id);

-- 12. 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_forum_posts_user_id ON forum_posts(user_id);
CREATE INDEX IF NOT EXISTS idx_forum_posts_created_at ON forum_posts(created_at);
CREATE INDEX IF NOT EXISTS idx_forum_posts_tags ON forum_posts USING gin(tags);
CREATE INDEX IF NOT EXISTS idx_forum_replies_post_id ON forum_replies(post_id);
CREATE INDEX IF NOT EXISTS idx_forum_replies_user_id ON forum_replies(user_id);
CREATE INDEX IF NOT EXISTS idx_forum_replies_parent_id ON forum_replies(parent_id);

-- 13. 验证表创建成功
SELECT '表创建完成，验证结果：' as status;
SELECT table_name, table_type 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name IN ('user_profiles', 'forum_posts', 'forum_replies');

-- 14. 显示备份表信息
SELECT '备份表信息：' as status;
SELECT table_name, table_type 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name LIKE '%_backup';