## 1. 服务端

- [x] 1.1 新增 `pets` migration up/down 与结构测试
- [x] 1.2 新增 `model.Pet`、请求 DTO、响应 DTO
- [x] 1.3 新增 `repo.PetRepo`：Create/ListByUser/FindByUserAndID/Update
- [x] 1.4 新增 `service.PetService`：校验、创建、列表、详情、更新、状态
- [x] 1.5 新增 `api.PetHandler` 并注册 `/api/v1/pets` 路由
- [x] 1.6 覆盖 service/api happy path 与错误路径测试

## 2. Flutter App

- [x] 2.1 新增宠物模型与 repository
- [x] 2.2 新增 pet controller：加载、创建、更新、选择第一只宠物
- [x] 2.3 新增宠物创建页
- [x] 2.4 改造主页为宠物主页壳，隐藏旧联系人主入口
- [x] 2.5 补充中文文案
- [x] 2.6 覆盖 controller 和主页关键 widget 测试

## 3. 验收

- [x] 3.1 `openspec validate add-pet-profile --strict` 通过
- [x] 3.2 等价服务端测试通过：`go test . -count=1`、`go test ./internal/api -count=1`、`go test ./internal/service -count=1`
- [x] 3.3 `flutter test` 和 `flutter analyze` 通过

