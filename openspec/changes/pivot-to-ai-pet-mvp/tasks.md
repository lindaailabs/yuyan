## 1. 文档与规格

- [x] 1.1 更新根 README / README-zh：产品定位改为 AI 宠物应用
- [x] 1.2 更新 docs/LLM_DEV_GUIDE.md：产品边界、架构分层、AI Gateway、记忆、成长、权益、数据要求
- [x] 1.3 更新 docs/MILESTONES.md：一期路线改为陪伴闭环
- [x] 1.4 更新 openspec/config.yaml：新任务上下文改为 AI 宠物主线
- [x] 1.5 更新 apps/app/README.md 与 pubspec 描述
- [x] 1.6 新增本 OpenSpec change，沉淀定位切换决策

## 2. 后续实现准备

- [ ] 2.1 新建 `add-pet-profile` change：pets 表、宠物创建、宠物主页
- [ ] 2.2 新建 `add-ai-gateway-pet-chat` change：AI Gateway mock、文本对话、消息持久化
- [ ] 2.3 新建 `add-pet-memory` change：记忆抽取、召回、展示、删除
- [ ] 2.4 新建 `add-pet-growth` change：成长规则、状态变化、成长事件
- [ ] 2.5 新建 `add-entitlement-analytics` change：权益、用量、埋点、成本统计
- [ ] 2.6 评估旧 contacts/friendship 入口：隐藏、保留或迁移

## 3. 验收

- [ ] 3.1 后续任务不再默认扩展好友/人与人 IM 主线
- [ ] 3.2 新增业务变更均引用 docs/LLM_DEV_GUIDE.md v2.0 对应章节
- [ ] 3.3 每个 AI 调用相关任务均支持 mock 测试，不依赖真实模型服务
