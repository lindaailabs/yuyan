## ADDED Requirements

### Requirement: 通讯录页
App SHALL 提供通讯录页（路由 `/contacts`，仅 AuthReady 可达）：进入时拉取 `GET /friends` 并按结交时间升序展示好友（头像、昵称、脱敏手机号）；空态 MUST 展示引导文案；加载失败 MUST 展示可重试入口。好友数据不做本地持久化，以服务端为唯一事实源。

#### Scenario: 展示好友
- **WHEN** 已有 2 位好友的用户进入通讯录
- **THEN** 依序展示 2 张好友卡片（AvatarWidget + 昵称 + 脱敏手机号）

#### Scenario: 空态
- **WHEN** 无好友用户进入通讯录
- **THEN** 展示空态引导文案与「去搜索」入口

### Requirement: 好友申请列表页
App SHALL 提供申请列表页（路由 `/requests`，仅 AuthReady 可达）：进入时拉取 `GET /friends/requests`，每条申请展示申请人头像/昵称/脱敏手机号与「同意/拒绝」按钮；点击同意 MUST 调用 accept 并在成功后从列表移除该条、同步通讯录数据；点击拒绝 MUST 调用 reject 并成功后从列表移除；操作失败 MUST 以 snackbar 提示文案且列表状态不破坏；重复点击操作中 MUST 置禁用防抖。

#### Scenario: 同意申请
- **WHEN** 用户点击某申请的「同意」且接口成功
- **THEN** 该申请从列表移除，通讯录刷新后可见新好友

#### Scenario: 拒绝申请
- **WHEN** 用户点击某申请的「拒绝」且接口成功
- **THEN** 该申请从列表移除，通讯录不变

#### Scenario: 操作失败
- **WHEN** accept/reject 接口返回错误（如 2106）
- **THEN** snackbar 展示失败文案，列表保持原状

### Requirement: 搜索页加好友入口
搜索结果卡片 SHALL 启用「加好友」按钮（W2 占位转真实）：点击 MUST 调用 `POST /friends/requests`；成功后按钮 MUST 变为「已申请」禁用态；业务错误（已是好友 2104 / 重复申请 2103 / 反向待处理 2105 等）MUST 以 snackbar 展示服务端 msg。

#### Scenario: 发送申请成功
- **WHEN** 用户在搜索结果点击「加好友」且接口返回 code=0
- **THEN** 按钮变为「已申请」禁用态

#### Scenario: 业务错误提示
- **WHEN** 目标已是好友时点击「加好友」
- **THEN** snackbar 展示「你们已经是好友」类服务端文案

### Requirement: 主页入口
主页 SHALL 提供通讯录与好友申请两个入口（图标按钮），分别跳转 `/contacts` 与 `/requests`；文案 MUST 收敛于 core/l10n。

#### Scenario: 入口跳转
- **WHEN** AuthReady 用户点击主页通讯录图标
- **THEN** 路由跳转到 /contacts
