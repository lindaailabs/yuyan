import 'package:yuyan_app/data/remote/api_client.dart';
import 'package:yuyan_app/data/remote/api_exception.dart';

/// ApiClient 假实现：按 "METHOD path" 路由到可编排的 handler（不触网）。
/// 供 controller 单测与需要 ProviderScope 的 widget 测试共用。
class FakeApiClient implements ApiClient {
  final Map<String, Future<Map<String, dynamic>> Function()> handlers = {};

  Future<Map<String, dynamic>> _run(String method, String path) {
    final h = handlers['$method $path'];
    if (h == null) {
      throw ApiException(500, 'no handler: $method $path');
    }
    return h();
  }

  @override
  Future<Map<String, dynamic>> get(String path,
          {Map<String, dynamic>? query}) =>
      _run('GET', path);

  @override
  Future<List<dynamic>> getList(String path,
      {Map<String, dynamic>? query}) async {
    return (await _run('GET', path))['items'] as List<dynamic>;
  }

  @override
  Future<Map<String, dynamic>> post(String path,
          {Map<String, dynamic>? body}) =>
      _run('POST', path);

  @override
  Future<Map<String, dynamic>> put(String path,
          {Map<String, dynamic>? body}) =>
      _run('PUT', path);
}
