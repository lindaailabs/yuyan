import 'package:dio/dio.dart';

import '../../core/auth/token_storage.dart';
import 'api_exception.dart';
import 'api_client.dart';

/// 基于 Dio 的 ApiClient 实现（guide §5.3）：
/// - 统一包裹解析：{code, msg, data}，code != 0 抛 [ApiException]
/// - token 注入：非 /auth/* 请求自动携带 Bearer access
/// - 401 刷新：refresh 一次 → 重放原请求；刷新失败触发 onSessionExpired
class DioApiClient implements ApiClient {
  DioApiClient({
    required String baseUrl,
    required TokenStorage tokenStorage,
    this.onSessionExpired,
    Dio? dio,
    Dio? refreshDio,
  }) : _storage = tokenStorage {
    _dio = dio ??
        Dio(BaseOptions(
          baseUrl: baseUrl,
          connectTimeout: const Duration(seconds: 5),
          receiveTimeout: const Duration(seconds: 10),
        ));
    // refresh 用独立裸 Dio（无拦截器，避免 401 递归）；测试可注入 mock adapter。
    _refreshDio = refreshDio ?? Dio(BaseOptions(baseUrl: baseUrl));
    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: _onRequest,
      onError: _onError,
    ));
  }

  final TokenStorage _storage;

  /// 会话彻底失效（refresh 也失败）时回调（AuthController.forceLogout）。
  final void Function()? onSessionExpired;

  late final Dio _dio;
  late final Dio _refreshDio;

  /// 匿名端点（登录三件套）：不带 Bearer，也不触发刷新。
  static const _anonymousPrefixes = ['/auth/sms-code', '/auth/login', '/auth/refresh'];

  void _onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final anonymous = _anonymousPrefixes.any((p) => options.path.startsWith(p));
    if (!anonymous) {
      // 异步注入 token（onRequest 回调内支持 Future）。
      _storage.readAccess().then((token) {
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      }).catchError((Object e) {
        handler.next(options); // 读失败按匿名放行，由服务端 401 兜底
      });
      return;
    }
    handler.next(options);
  }

  Future<void> _onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final options = err.requestOptions;
    final isAuthPath = _anonymousPrefixes.any((p) => options.path.startsWith(p));
    final notRefreshed = options.extra['_retried'] != true;

    // 仅业务端点的 401 尝试刷新（/auth/login 的 1002 不属于会话过期场景）。
    if (err.response?.statusCode != 401 || isAuthPath || !notRefreshed) {
      handler.next(err);
      return;
    }

    final refresh = await _storage.readRefresh();
    if (refresh == null || refresh.isEmpty) {
      _notifySessionExpired();
      handler.next(err);
      return;
    }

    // refresh 一次：用独立裸 Dio 避免拦截器递归。
    try {
      final resp = await _refreshDio.post(
        '/auth/refresh',
        data: {'refresh_token': refresh},
      );
      final data = _unwrap(resp.data);
      final access = data['access_token'] as String;
      final newRefresh = data['refresh_token'] as String;
      await _storage.write(access: access, refresh: newRefresh);

      // 重放原请求。
      options.extra['_retried'] = true;
      options.headers['Authorization'] = 'Bearer $access';
      final replayed = await _dio.fetch(options);
      handler.resolve(replayed);
    } catch (_) {
      _notifySessionExpired();
      handler.next(err);
    }
  }

  void _notifySessionExpired() {
    onSessionExpired?.call();
  }

  /// 统一包裹解析：code != 0 → ApiException。
  Map<String, dynamic> _unwrap(dynamic body) {
    if (body is! Map<String, dynamic>) {
      throw const ApiException(5001, '响应格式错误');
    }
    final code = body['code'] as int? ?? 5001;
    final msg = body['msg'] as String? ?? '未知错误';
    if (code != 0) {
      throw ApiException(code, msg);
    }
    final data = body['data'];
    if (data is Map<String, dynamic>) {
      return data;
    }
    return const {}; // data 允许为 null（如 void 场景）
  }

  /// 统一包裹解析（数组形态）：code != 0 → ApiException。
  List<dynamic> _unwrapList(dynamic body) {
    if (body is! Map<String, dynamic>) {
      throw const ApiException(5001, '响应格式错误');
    }
    final code = body['code'] as int? ?? 5001;
    final msg = body['msg'] as String? ?? '未知错误';
    if (code != 0) {
      throw ApiException(code, msg);
    }
    final data = body['data'];
    if (data is List) {
      return data;
    }
    return const []; // data 允许为 null
  }

  @override
  Future<Map<String, dynamic>> get(String path,
      {Map<String, dynamic>? query}) async {
    final resp = await _dio.get(path, queryParameters: query);
    return _unwrap(resp.data);
  }

  @override
  Future<List<dynamic>> getList(String path,
      {Map<String, dynamic>? query}) async {
    final resp = await _dio.get(path, queryParameters: query);
    return _unwrapList(resp.data);
  }

  @override
  Future<Map<String, dynamic>> post(String path,
      {Map<String, dynamic>? body}) async {
    final resp = await _dio.post(path, data: body);
    return _unwrap(resp.data);
  }

  @override
  Future<Map<String, dynamic>> put(String path,
      {Map<String, dynamic>? body}) async {
    final resp = await _dio.put(path, data: body);
    return _unwrap(resp.data);
  }

  @override
  Future<Map<String, dynamic>> delete(String path) async {
    final resp = await _dio.delete(path);
    return _unwrap(resp.data);
  }
}
