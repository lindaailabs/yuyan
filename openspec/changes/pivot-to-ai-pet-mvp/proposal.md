## Why

Yuyan 当前仓库和文档仍带有早期“即时通讯 App”主线，但新的产品定位已调整为 AI 宠物应用：一只有记忆、会成长、能对话的虚拟宠物。继续按好友关系、多人单聊、离线补发推进，会把研发资源消耗在非核心差异化上。

本变更把一期 MVP 的验收标准切换为“陪伴闭环”：登录、创建宠物、文本对话、记忆、成长、用量统计和权益地基。

## What Changes

- 更新项目上下文：README、LLM_DEV_GUIDE、MILESTONES、OpenSpec context 改为 AI 宠物方向。
- 明确旧 IM 能力的处理方式：账号、消息可靠性、WebSocket、分页、本地缓存可复用；好友、通讯录、人与人聊天不再作为一期主线。
- 新增一期核心能力定义：
  - `pet-profile`: 宠物档案、外观、persona、状态
  - `pet-conversation`: 用户与宠物文本对话、消息持久化、游标分页
  - `ai-gateway`: 模型调用、prompt 版本、上下文拼装、用量统计、兜底
  - `pet-memory`: 记忆抽取、召回、展示、删除
  - `pet-growth`: 亲密度、心情、等级、成长事件
  - `entitlement`: 免费/订阅权益和额度模型
  - `analytics`: AI 成本与关键行为事件

## Capabilities

### New Capabilities

- `ai-pet-mvp`: 一期 AI 宠物陪伴闭环的产品与技术规格。

### Modified Capabilities

- `app-skeleton`: 后续页面主线从联系人/好友切换为宠物主页、聊天、记忆、成长、权益。
- `server-skeleton`: 后续业务域优先 pet、memory、ai、entitlement、analytics。
- `protocol-schema`: 早期 IM v1 schema 暂保留；AI 宠物流式/实时协议需另开 v2 规格后再实现。

## Impact

- 代码：本变更仅更新文档与规格，不直接修改业务实现。
- 数据库：后续将新增 pets、pet_conversations、pet_messages、pet_memories、pet_growth_events、entitlements、usage/event 相关表。
- 风险：现有好友/联系人代码会成为遗留资产。短期保留，不主动扩展；如影响新主线体验，再单独迁移或隐藏入口。
- 新依赖：本变更不引入新依赖。
