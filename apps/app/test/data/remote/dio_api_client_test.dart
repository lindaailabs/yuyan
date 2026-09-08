import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/auth/token_storage.dart';
import 'package:yuyan_app/data/remote/api_exception.dart';
import 'package:yuyan_app/data/remote/dio_api_client.dart';

/// 按「方法 + 路径」路由的 mock adapter：记录请求头供断言。
class _MockAdapter implements HttpClientAdapter {
  final List<RequestOptions> requests = [];
  final Map<String, Future<ResponseBody> Function(RequestOptions)> routes = {};

  void on(String method, String path,
      Future<ResponseBody> Function(RequestOptions) responder) {
    routes['$method $path'] = responder;
  }

  @override
  Future<ResponseBody> fetch(RequestOptions options,
      Stream<Uint8List>? requestStream, Future<void>? cancelFuture) async {
    requests.add(options);
    final key = '${options.method} ${options.path}';
    final responder = routes[key];
    if (responder == null) {
      throw StateError('no route: $key');
    }
    return responder(options);
  }

  @override
  void close({bool force = false}) {}
}

ResponseBody _json(Object body, int status) => ResponseBody.fromString(
      jsonEncode(body),
      status,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );

Map<String, dynamic> _me = {
  'id': 1,
  'phone': '13811112222',
  'nickname': '语燕',
  'avatar_id': 1,
  'created_at': 1788858000,
};

void main() {
  const baseUrl = 'http://localhost:9999/api/v1';

  late _MockAdapter adapter;
  late MemoryTokenStorage storage;
  late List<String> expiredCalls;

  DioApiClient newClient() {
    final dio = Dio(BaseOptions(baseUrl: baseUrl))..httpClientAdapter = adapter;
    final refreshDio = Dio(BaseOptions(baseUrl: baseUrl))
      ..httpClientAdapter = adapter;
    return DioApiClient(
      baseUrl: baseUrl,
      tokenStorage: storage,
      onSessionExpired: () => expiredCalls.add('expired'),
      dio: dio,
      refreshDio: refreshDio,
    );
  }

  setUp(() {
    adapter = _MockAdapter();
    storage = MemoryTokenStorage();
    expiredCalls = [];
  });

  test('token 注入：非 /auth 端点携带 Bearer access，匿名端点不带', () async {
    await storage.write(access: 'old-a', refresh: 'old-r');
    adapter.on('GET', '/users/me', (o) async {
      expect(o.headers['Authorization'], 'Bearer old-a');
      return _json({'code': 0, 'msg': 'ok', 'data': _me}, 200);
    });
    adapter.on('POST', '/auth/login', (o) async {
      expect(o.headers.containsKey('Authorization'), isFalse);
      return _json({'code': 0, 'msg': 'ok', 'data': {}}, 200);
    });

    final client = newClient();
    final me = await client.get('/users/me');
    expect(me['nickname'], '语燕');
    await client.post('/auth/login', body: {'phone': '13811112222', 'code': '1'});
  });

  test('包裹解析：code != 0 抛 ApiException（code/msg 透传）', () async {
    adapter.on('GET', '/users/me',
        (o) async => _json({'code': 2001, 'msg': '验证码错误'}, 200));

    final client = newClient();
    await expectLater(
      client.get('/users/me'),
      throwsA(isA<ApiException>()
          .having((e) => e.code, 'code', 2001)
          .having((e) => e.msg, 'msg', '验证码错误')),
    );
  });

  test('401 → refresh 一次 → 重放原请求（新 token），不触发 onSessionExpired',
      () async {
    await storage.write(access: 'old-a', refresh: 'old-r');
    var meCalls = 0;
    adapter.on('GET', '/users/me', (o) async {
      meCalls++;
      if (meCalls == 1) {
        expect(o.headers['Authorization'], 'Bearer old-a');
        return _json({'code': 1002, 'msg': '凭证无效'}, 401);
      }
      // 重放请求应携带刷新后的 token。
      expect(o.headers['Authorization'], 'Bearer new-a');
      return _json({'code': 0, 'msg': 'ok', 'data': _me}, 200);
    });
    adapter.on('POST', '/auth/refresh', (o) async {
      expect(o.data, {'refresh_token': 'old-r'});
      return _json({
        'code': 0,
        'msg': 'ok',
        'data': {'access_token': 'new-a', 'refresh_token': 'new-r'},
      }, 200);
    });

    final client = newClient();
    final me = await client.get('/users/me');
    expect(me['nickname'], '语燕');
    expect(await storage.readAccess(), 'new-a');
    expect(await storage.readRefresh(), 'new-r');
    expect(expiredCalls, isEmpty);
    // 只重放一次：/users/me 共 2 次、refresh 共 1 次。
    expect(meCalls, 2);
    expect(
      adapter.requests.where((r) => r.path == '/auth/refresh').length,
      1,
    );
  });

  test('401 且 refresh 失败 → onSessionExpired，原异常上抛', () async {
    await storage.write(access: 'old-a', refresh: 'old-r');
    adapter.on('GET', '/users/me',
        (o) async => _json({'code': 1002, 'msg': '凭证无效'}, 401));
    adapter.on('POST', '/auth/refresh',
        (o) async => _json({'code': 1002, 'msg': 'refresh 已失效'}, 401));

    final client = newClient();
    await expectLater(
      client.get('/users/me'),
      throwsA(isA<DioException>()),
    );
    expect(expiredCalls, ['expired']);
  });

  test('401 且本地无 refresh token → 直接 onSessionExpired，不发 refresh',
      () async {
    adapter.on('GET', '/users/me',
        (o) async => _json({'code': 1002, 'msg': '凭证无效'}, 401));

    final client = newClient();
    await expectLater(
      client.get('/users/me'),
      throwsA(isA<DioException>()),
    );
    expect(expiredCalls, ['expired']);
    expect(
      adapter.requests.where((r) => r.path == '/auth/refresh'),
      isEmpty,
    );
  });

  test('getList：data 为数组时解包返回列表', () async {
    await storage.write(access: 'a', refresh: 'r');
    adapter.on('GET', '/users/search', (o) async {
      expect(o.uri.queryParameters['q'], '13811112222');
      return _json({
        'code': 0,
        'msg': 'ok',
        'data': [
          {'id': 2, 'phone': '138****2222', 'nickname': '小明', 'avatar_id': 2},
        ],
      }, 200);
    });

    final client = newClient();
    final items = await client.getList('/users/search', query: {'q': '13811112222'});
    expect(items, hasLength(1));
    expect((items.first as Map<String, dynamic>)['nickname'], '小明');
  });
}
