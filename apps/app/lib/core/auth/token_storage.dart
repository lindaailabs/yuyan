import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Token 安全存储（guide §5.3：token 只存 flutter_secure_storage，禁 SharedPreferences/明文）。
abstract class TokenStorage {
  Future<String?> readAccess();
  Future<String?> readRefresh();

  /// 写入双 token。
  Future<void> write({required String access, required String refresh});

  /// 清空（登出/会话失效）。
  Future<void> clear();
}

/// 生产实现：flutter_secure_storage（Android Keystore / iOS Keychain）。
class SecureTokenStorage implements TokenStorage {
  SecureTokenStorage({FlutterSecureStorage? storage})
      : _storage = storage ?? const FlutterSecureStorage();

  static const _keyAccess = 'auth.access_token';
  static const _keyRefresh = 'auth.refresh_token';

  final FlutterSecureStorage _storage;

  @override
  Future<String?> readAccess() => _storage.read(key: _keyAccess);

  @override
  Future<String?> readRefresh() => _storage.read(key: _keyRefresh);

  @override
  Future<void> write({required String access, required String refresh}) async {
    await _storage.write(key: _keyAccess, value: access);
    await _storage.write(key: _keyRefresh, value: refresh);
  }

  @override
  Future<void> clear() async {
    await _storage.delete(key: _keyAccess);
    await _storage.delete(key: _keyRefresh);
  }
}

/// 内存实现（单测用，避免依赖平台插件）。
class MemoryTokenStorage implements TokenStorage {
  String? _access;
  String? _refresh;

  @override
  Future<String?> readAccess() async => _access;

  @override
  Future<String?> readRefresh() async => _refresh;

  @override
  Future<void> write({required String access, required String refresh}) async {
    _access = access;
    _refresh = refresh;
  }

  @override
  Future<void> clear() async {
    _access = null;
    _refresh = null;
  }
}
