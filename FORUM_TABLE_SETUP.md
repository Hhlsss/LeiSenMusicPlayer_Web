# 音乐论坛数据库表设置指南

## 概述
本文档指导您在Supabase中创建音乐论坛所需的数据库表结构。

## 表结构SQL脚本

### 1. 论坛帖子表 (forum_posts)
```sql
-- 创建论坛帖子表
CREATE TABLE IF NOT EXISTS forum_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES auth.users(id),
    user_nickname VARCHAR(100) NOT NULL,
    music_reference JSONB,
    tags VARCHAR[] DEFAULT '{}',
    view_count INTEGER DEFAULT 0,
    reply_count INTEGER DEFAULT 0,
    is_pinned BOOLEAN DEFAULT false,
    is_locked BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 启用行级安全
ALTER TABLE forum_posts ENABLE ROW LEVEL SECURITY;

-- 创建策略：所有用户都可以查看帖子
CREATE POLICY "任何人都可以查看帖子" ON forum_posts
FOR SELECT USING (true);

-- 创建策略：认证用户可以创建帖子
CREATE POLICY "认证用户可以创建帖子" ON forum_posts
FOR INSERT WITH CHECK (auth.uid() = user_id);

-- 创建策略：用户只能编辑自己的帖子
CREATE POLICY "用户只能编辑自己的帖子" ON forum_posts
FOR UPDATE USING (auth.uid() = user_id);

-- 创建策略：用户只能删除自己的帖子
CREATE POLICY "用户只能删除自己的帖子" ON forum_posts
FOR DELETE USING (auth.uid() = user_id);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_forum_posts_created_at ON forum_posts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_forum_posts_user_id ON forum_posts(user_id);
CREATE INDEX IF NOT EXISTS idx_forum_posts_tags ON forum_posts USING GIN(tags);
```

### 2. 论坛回复表 (forum_replies)
```sql
-- 创建论坛回复表
CREATE TABLE IF NOT EXISTS forum_replies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES forum_posts(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES auth.users(id),
    user_nickname VARCHAR(100) NOT NULL,
    parent_id UUID REFERENCES forum_replies(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 启用行级安全
ALTER TABLE forum_replies ENABLE ROW LEVEL SECURITY;

-- 创建策略：所有用户都可以查看回复
CREATE POLICY "任何人都可以查看回复" ON forum_replies
FOR SELECT USING (true);

-- 创建策略：认证用户可以创建回复
CREATE POLICY "认证用户可以创建回复" ON forum_replies
FOR INSERT WITH CHECK (auth.uid() = user_id);

-- 创建策略：用户只能编辑自己的回复
CREATE POLICY "用户只能编辑自己的回复" ON forum_replies
FOR UPDATE USING (auth.uid() = user_id);

-- 创建策略：用户只能删除自己的回复
CREATE POLICY "用户只能删除自己的回复" ON forum_replies
FOR DELETE USING (auth.uid() = user_id);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_forum_replies_post_id ON forum_replies(post_id);
CREATE INDEX IF NOT EXISTS idx_forum_replies_user_id ON forum_replies(user_id);
CREATE INDEX IF NOT EXISTS idx_forum_replies_parent_id ON forum_replies(parent_id);
CREATE INDEX IF NOT EXISTS idx_forum_replies_created_at ON forum_replies(created_at);
```

## 执行步骤

### 方法1：通过Supabase SQL编辑器执行
1. 登录 [Supabase控制台](https://supabase.com/dashboard)
2. 选择您的项目
3. 进入左侧菜单的 "SQL Editor"
4. 将上面的SQL脚本复制到编辑器中
5. 点击 "Run" 执行

### 方法2：通过Table Editor创建
1. 登录 Supabase控制台
2. 选择您的项目
3. 进入 "Table Editor"
4. 点击 "Create a new table"
5. 按照上面的表结构手动创建字段

## 表结构说明

### forum_posts (论坛帖子表)
- `id`: 主键，UUID格式
- `title`: 帖子标题
- `content`: 帖子内容
- `user_id`: 发帖用户ID
- `user_nickname`: 用户昵称
- `music_reference`: 音乐引用信息（JSON格式）
- `tags`: 标签数组
- `view_count`: 查看次数
- `reply_count`: 回复数量
- `is_pinned`: 是否置顶
- `is_locked`: 是否锁定
- `created_at`: 创建时间
- `updated_at`: 更新时间

### forum_replies (论坛回复表)
- `id`: 主键，UUID格式
- `post_id`: 关联的帖子ID
- `content`: 回复内容
- `user_id`: 回复用户ID
- `user_nickname`: 用户昵称
- `parent_id`: 父回复ID（支持嵌套回复）
- `created_at`: 创建时间
- `updated_at`: 更新时间

## 安全策略
- 所有用户都可以查看帖子和回复
- 只有认证用户可以创建帖子和回复
- 用户只能编辑和删除自己的内容

## 索引优化
- 按创建时间倒序索引，便于显示最新帖子
- 用户ID索引，便于查询用户发帖
- 标签GIN索引，支持标签搜索
- 帖子ID索引，便于查询回复

创建完成后，您的音乐论坛数据库就准备好了！