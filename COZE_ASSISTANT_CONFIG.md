# 智能音乐助手配置说明

## 🎵 智能音乐助手配置信息

### Coze智能体链接
- **智能体URL**: https://www.coze.cn/space/7559750439882375218/bot/7567959568312369215
- **空间ID**: `7559750439882375218`
- **智能体ID**: `7567959568312369215`

### API配置详情
- **API Token**: `pat_1mZRTEMafM2v3ZjLZa2ioCoNRBp86Dh3bKl3X4667qtYgB0FLbOVW90fdkppYsaq`
- **API端点**: `https://api.coze.cn/v1/chat/completions`

### 功能特性

#### 1. 界面设计
- 位于首页轮播图右侧区域
- 与网站红色主题协调一致
- 响应式设计，适配不同屏幕尺寸

#### 2. 核心功能
- ✅ 实时对话交互
- ✅ 加载状态显示
- ✅ 键盘支持（Enter键发送）
- ✅ 自动滚动到底部
- ✅ 输入验证和状态管理

#### 3. 智能体专业能力
- 🎶 歌曲推荐和歌单管理
- 🎵 音乐风格分析和流派识别
- 📚 音乐知识解答
- 🔍 音乐搜索和发现
- 💡 音乐制作建议

### 技术实现

#### API请求参数
```javascript
{
  bot_id: '7567959568312369215',
  conversation_id: 'music_assistant_session',
  user: 'music_web_user',
  query: '用户输入的消息',
  stream: false,
  custom_variables: {
    platform: 'music_web',
    version: '1.0'
  }
}
```

#### 响应格式适配
支持多种API响应格式：
1. **Coze原生格式** (`data.messages[0].content`)
2. **OpenAI兼容格式** (`data.choices[0].message.content`)
3. **直接内容格式** (`data.content`)

### 错误处理
- 🔄 网络异常自动重试
- 📱 友好的错误提示信息
- 🔍 详细的调试日志

### 用户体验优化
- 🔍 输入框自动聚焦
- ⚡ 发送按钮状态实时更新
- 📱 移动端适配优化
- 🎨 加载动画和状态指示

## 🔧 维护说明

### 配置更新
如需更新智能体配置，请修改 `web/index.html` 文件中的以下变量：

```javascript
const BOT_ID = '7567959568312369215';
const SPACE_ID = '7559750439882375218';
const COZE_API_TOKEN = '您的API Token';
```

### API端点变更
如果Coze API端点发生变化，请更新：
```javascript
const COZE_API_URL = '新的API端点';
```

### 功能扩展建议
1. **会话持久化** - 保存对话历史
2. **多语言支持** - 支持英文等其他语言
3. **语音交互** - 集成语音输入输出
4. **个性化推荐** - 基于用户历史优化推荐

## 📞 技术支持

如遇到技术问题，请检查：
1. API Token是否有效
2. 网络连接是否正常
3. 智能体是否在Coze平台正常运行
4. 浏览器控制台错误信息

---
*最后更新: 2025-11-02*