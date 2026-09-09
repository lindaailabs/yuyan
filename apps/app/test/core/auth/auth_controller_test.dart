import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/auth/auth_controller.dart';
import 'package:yuyan_app/core/auth/auth_state.dart';
import 'package:yuyan_app/core/auth/token_storage.dart';
import 'package:yuyan_app/data/remote/api_exception.dart';
import 'package:yuyan_app/data/repository/auth_repository.dart';

import '../../support/fake_api_client.dart';

Map<String, dynamic> _meJson({String? nickname}) => {
      'id': 1,
      'phone': '13811112222',
      'nickname': nickname,
      'avatar_id': 1,
      'created_at': 1788858000,
    };

void main() {
  late FakeApiClient api;
  late MemoryTokenStorage storage;
  late AuthRepository repo;

  setUp(() {
    api = FakeApiClient();
    storage = MemoryTokenStorage();
    repo = AuthRepository(api);
  });

  AuthController newController() => AuthController(repo, storage);

  test('init：本地无 token → LoggedOut，不请求网络', () async {
    var meCalled = false;
    api.handlers['GET /users/me'] = () async {
      meCalled = true;
      return _meJson(nickname: '语燕');
    };

    final c = newController();
    await c.init();
    expect(c.state, isA<AuthLoggedOut>());
    expect(meCalled, isFalse);
  });

  test('init：有 token 且 nickname 非空 → Ready', () async {
    await storage.write(access: 'a', refresh: 'r');
    api.handlers['GET /users/me'] =
        () async => _meJson(nickname: '语燕');

    final c = newController();
    await c.init();
    expect(c.state, isA<AuthReady>());
    expect((c.state as AuthReady).profile.nickname, '语燕');
  });

  test('init：有 token 但 nickname 为 null → Onboarding', () async {
    await storage.write(access: 'a', refresh: 'r');
    api.handlers['GET /users/me'] = () async => _meJson();

    final c = newController();
    await c.init();
    expect(c.state, isA<AuthOnboarding>());
  });

  test('init：会话失效（ApiException）→ 清库回落 LoggedOut', () async {
    await storage.write(access: 'a', refresh: 'r');
    api.handlers['GET /users/me'] = () async =>
        throw const ApiException(1002, '凭证无效');

    final c = newController();
    await c.init();
    expect(c.state, isA<AuthLoggedOut>());
    expect(await storage.readAccess(), isNull);
    expect(await storage.readRefresh(), isNull);
  });

  test('login：成功后写双 token，并按 nickname 分流', () async {
    api.handlers['POST /auth/login'] = () async => {
          'access_token': 'new-a',
          'refresh_token': 'new-r',
          'expires_in': 7200,
        };
    api.handlers['GET /users/me'] = () async => _meJson();

    final c = newController();
    await c.login('13811112222', '123456');
    expect(c.state, isA<AuthOnboarding>());
    expect(await storage.readAccess(), 'new-a');
    expect(await storage.readRefresh(), 'new-r');
  });

  test('completeOnboarding：提交昵称+头像 → Ready', () async {
    api.handlers['PUT /users/me'] =
        () async => _meJson(nickname: '语燕')..['avatar_id'] = 3;

    final c = newController();
    await c.completeOnboarding('语燕', 3);
    expect(c.state, isA<AuthReady>());
    final profile = (c.state as AuthReady).profile;
    expect(profile.nickname, '语燕');
    expect(profile.avatarId, 3);
  });

  test('updateProfile：更新资料 → Ready 且返回新资料', () async {
    api.handlers['PUT /users/me'] =
        () async => _meJson(nickname: '新昵称')..['avatar_id'] = 5;

    final c = newController();
    final profile = await c.updateProfile(nickname: '新昵称', avatarId: 5);
    expect(profile.nickname, '新昵称');
    expect(c.state, isA<AuthReady>());
  });

  test('forceLogout/logout：清库回落 LoggedOut', () async {
    await storage.write(access: 'a', refresh: 'r');

    final c = newController();
    await c.logout();
    expect(c.state, isA<AuthLoggedOut>());
    expect(await storage.readAccess(), isNull);
  });
}
