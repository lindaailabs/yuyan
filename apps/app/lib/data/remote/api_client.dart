/// 远程网络层：REST API 客户端抽象（guide §3.4：widget 树禁止直接调用）。
/// 返回值为统一包裹中的 data 部分（实现负责解包与错误转换）。
abstract class ApiClient {
  Future<Map<String, dynamic>> get(String path, {Map<String, dynamic>? query});

  /// data 为 JSON 数组的端点（如用户搜索）。
  Future<List<dynamic>> getList(String path, {Map<String, dynamic>? query});

  Future<Map<String, dynamic>> post(String path, {Map<String, dynamic>? body});
  Future<Map<String, dynamic>> put(String path, {Map<String, dynamic>? body});
  Future<Map<String, dynamic>> delete(String path);
}
