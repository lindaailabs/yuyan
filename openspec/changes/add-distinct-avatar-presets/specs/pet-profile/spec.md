## MODIFIED Requirements

### Requirement: 宠物档案

系统 SHALL 支持当前登录用户创建、读取、更新自己的宠物档案。宠物头像/外观 SHALL 使用 12 个宠物预置头像（assets/pet_avatars/pet_avatar_1~12），且服务端 SHALL 接受 avatar_id 1~12。

#### Scenario: Create pet

- **WHEN** 已登录用户提交合法宠物名字、物种和 avatar_id 12
- **THEN** 创建成功，返回该宠物档案

#### Scenario: Reject invalid pet avatar

- **WHEN** 已登录用户提交 avatar_id 13 或更大值
- **THEN** 服务端 SHALL 返回宠物外观不存在业务错误
