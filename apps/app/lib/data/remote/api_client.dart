/// 远程网络层：REST API 客户端抽象（guide §3.4：widget 树禁止直接调用）。
/// W1 为占位接口；W2 实现 Dio 封装（baseURL /api/v1、token 注入、自动刷新）。
abstract class ApiClient {
  Future<Map<String, dynamic>> get(String path, {Map<String, dynamic>? query});
  Future<Map<String, dynamic>> post(String path, {Map<String, dynamic>? body});
  Future<Map<String, dynamic>> put(String path, {Map<String, dynamic>? body});
}
