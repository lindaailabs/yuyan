## Why

AI 宠物方向的一期闭环需要先让用户拥有一只可恢复的宠物。当前账号体系已经完成，但主页仍是早期 IM/联系人入口；缺少 pets 表、宠物档案接口和客户端宠物主页。

## What Changes

- 新增 `pets` 表：用户、名字、物种、头像/外观、persona、等级、亲密度、心情、创建/更新时间。
- 新增服务端宠物档案能力：创建宠物、查询我的宠物、查询详情、更新档案、查询状态。
- 新增 Flutter 宠物域：模型、repository、Riverpod 状态机、创建页、宠物主页壳。
- 调整 App 主页：从联系人/好友入口改为 AI 宠物第一屏；旧联系人入口不再作为主线展示。

## Capabilities

### New Capabilities

- `pet-profile`: 宠物创建、档案查询、状态查询和基础更新。

### Modified Capabilities

- `app-auth`: 登录完成后进入宠物主页。
- `app-skeleton`: 主页主线从 IM 占位切到宠物状态与创建入口。

## Impact

- 代码：apps/server 新增 pet model/repo/service/api/migration/tests；apps/app 新增 pet model/repository/controller/pages/tests。
- 数据库：新增 `pets` 表，不修改历史 migration。
- 新依赖：无。
- 风险：旧 contacts 入口仍存在但从主页隐藏，后续再决定删除或迁移。
