# app-skeleton Specification

## Purpose

Flutter 客户端骨架：路由壳、主题、drift 初始化、网络层与 repository 接口占位（apps/app，分层见 LLM_DEV_GUIDE.md §3.4）。为 W2+ 业务功能提供地基。

## Requirements

### Requirement: 应用启动与路由壳
Flutter 应用（package `yuyan_app`，applicationId/BundleID `com.yuyan.app`）SHALL 以 go_router 路由壳启动，包含登录页与主页两个占位路由；目录结构 MUST 遵循 LLM_DEV_GUIDE.md §3.4（core/data/features/shared）。

#### Scenario: 冷启动
- **WHEN** 在模拟器上启动应用
- **THEN** 应用 3 秒内显示占位主页，无报错

#### Scenario: 目录分层就位
- **WHEN** 检查 apps/app/lib 目录
- **THEN** core/（theme、l10n、router）、data/（remote、local、repository）、features/、shared/ 均存在且有对应文件

### Requirement: 状态管理与依赖方向
应用根节点 MUST 被 ProviderScope 包裹（Riverpod）；provider 层不得引用 widget，UI 不得直接调用 data/remote（Dio/WebSocket 客户端禁止在 widget 树出现）。

#### Scenario: ProviderScope 接入
- **WHEN** 检查 main.dart
- **THEN** runApp 外层包裹 ProviderScope，存在至少一个示例 provider

### Requirement: drift 本地数据库初始化
应用 SHALL 初始化 drift 数据库（data/local，版本 1 空迁移），启动时打开数据库并在关闭时正确释放；W1 不含任何业务表。

#### Scenario: 数据库打开与关闭
- **WHEN** 应用启动后进入主页再退出
- **THEN** drift 数据库成功打开，无迁移错误，退出无泄漏告警

### Requirement: 网络层与 repository 占位
data/remote SHALL 提供 Dio 客户端与 WebSocket 客户端的接口定义（空实现），data/repository SHALL 提供对应 repository 抽象占位；后续业务功能在这些接口内扩展。

#### Scenario: 占位接口存在
- **WHEN** 检查 data/remote 与 data/repository
- **THEN** 存在接口文件且被 repository 层引用，widget 层 import 列表中无 dio/web_socket_channel

### Requirement: 主题与文案规范
应用主题 MUST 收敛于 core/theme；一期可见文案 MUST 收敛于 core/l10n（中文常量），禁止硬编码在 widget 中。

#### Scenario: 占位页文案来源
- **WHEN** 检查占位页面源码
- **THEN** 所有用户可见字符串引用 core/l10n 常量
