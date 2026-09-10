## MODIFIED Requirements

### Requirement: 首登引导

新用户（nickname=null）登录后 SHALL 进入引导页：从 8 个用户预置头像（assets/user_avatars/user_avatar_1~8）选择一个，并输入 1~20 字符昵称，提交调用 PUT /users/me；成功后进入主页。头像未选择时提交按钮禁用。

#### Scenario: 设置资料成功

- **WHEN** 选择头像 3 并输入昵称提交成功
- **THEN** 路由跳转主页，资料页可见新昵称与头像

### Requirement: 我的资料页

资料页 SHALL 展示当前用户昵称/头像/手机号，支持编辑昵称与切换 8 个用户预置头像（提交 PUT /users/me），保存后即时生效。

#### Scenario: 切换头像

- **WHEN** 选择头像 5 保存成功
- **THEN** 页面与后续用户资料展示头像 5
