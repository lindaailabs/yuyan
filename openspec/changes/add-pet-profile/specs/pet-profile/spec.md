# pet-profile Specification

## ADDED Requirements

### Requirement: Pet Creation

系统 SHALL 允许已登录用户创建 AI 宠物档案。

#### Scenario: Create pet

- WHEN 已登录用户提交合法宠物名字、物种和头像 id
- THEN 系统 SHALL 创建归属于该用户的宠物
- AND SHALL 初始化等级为 1、亲密度为 0、心情为默认值

#### Scenario: Reject invalid pet name

- WHEN 已登录用户提交空名字或超过 20 字符的名字
- THEN 系统 SHALL 返回业务错误

### Requirement: User-Scoped Pet Access

系统 SHALL 仅允许用户访问自己的宠物。

#### Scenario: Read own pet

- WHEN 用户查询自己创建的宠物
- THEN 系统 SHALL 返回宠物档案和状态字段

#### Scenario: Reject cross-user access

- WHEN 用户查询或更新其他用户的宠物 id
- THEN 系统 SHALL 返回宠物不存在或不可访问错误

### Requirement: Pet Home Surface

客户端 SHALL 在登录完成后展示 AI 宠物主页，而不是早期联系人/好友主页。

#### Scenario: No pet yet

- WHEN 当前用户没有宠物
- THEN App SHALL 展示创建宠物入口

#### Scenario: Existing pet

- WHEN 当前用户已有宠物
- THEN App SHALL 展示宠物名字、心情、等级和亲密度
