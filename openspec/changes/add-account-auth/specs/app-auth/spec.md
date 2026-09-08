## ADDED Requirements

### Requirement: 登录页
登录页 SHALL 提供手机号输入、验证码输入与"获取验证码"按钮；点击获取后展示响应中的图形验证码图片（base64）；响应无 `captcha_image` 字段时 MUST 显示"验证码已发送，请查收短信"兼容态。验证码图片支持点击刷新。手机号格式错误时禁用提交并给出提示（文案出自 core/l10n）。

#### Scenario: 图形验证码展示
- **WHEN** 获取验证码成功且响应含 captcha_image
- **THEN** 页面显示验证码图片

#### Scenario: 短信兼容态
- **WHEN** 响应无 captcha_image
- **THEN** 页面显示等待短信提示，验证码输入框可用

### Requirement: 首登引导页
新用户（nickname=null）登录后 SHALL 进入引导页：从 8 个预置头像（assets/avatars/avatar_1~8）选择一个，并输入 1~20 字符昵称，提交调用 PUT /users/me；成功后进入主页。头像未选择时提交按钮禁用。

#### Scenario: 完成引导
- **WHEN** 选择头像 3 并输入昵称提交成功
- **THEN** 路由跳转主页，资料页可见新昵称与头像

#### Scenario: 未选头像
- **WHEN** 未选头像点击提交
- **THEN** 提交按钮禁用

### Requirement: 路由三态守卫
App 路由 SHALL 按 auth 状态分流：无 token → 登录页；有 token 且 nickname=null → 引导页；否则 → 主页。启动时从安全存储恢复 token，恢复期间显示加载态。

#### Scenario: 冷启动三态
- **WHEN** 分别在无 token / 有 token 未引导 / 有 token 已引导 三种持久化状态下启动
- **THEN** 分别落在 登录页 / 引导页 / 主页

### Requirement: Token 生命周期
App SHALL 将 access/refresh token 持久化到 flutter_secure_storage；Dio 拦截器在 401 时自动以 refresh token 换新并重放原请求；refresh 失效时清除本地 token 并回到登录页。并发请求 401 时 MUST 只触发一次 refresh。

#### Scenario: 自动刷新
- **WHEN** access 过期后发起业务请求
- **THEN** 拦截器静默刷新并重放，业务层无感

#### Scenario: 登出态
- **WHEN** refresh token 也过期（刷新返回 1002）
- **THEN** 清空 token，路由回登录页

### Requirement: 我的资料页
资料页 SHALL 展示当前用户 昵称/头像/手机号，支持编辑昵称与切换 8 个预置头像（提交 PUT /users/me），保存后即时生效。

#### Scenario: 切换头像
- **WHEN** 选择头像 5 保存成功
- **THEN** 页面与后续会话显示头像 5

### Requirement: 用户搜索页
搜索页 SHALL 提供手机号输入与搜索按钮，展示命中结果（昵称/头像/脱敏手机号）；未命中显示空态；命中项仅展示（"添加好友"按钮预留禁用，W3 启用）。

#### Scenario: 搜索命中
- **WHEN** 输入已注册手机号并搜索
- **THEN** 显示该用户卡片

#### Scenario: 空态
- **WHEN** 输入未注册手机号
- **THEN** 显示"未找到用户"空态（文案出自 core/l10n）
