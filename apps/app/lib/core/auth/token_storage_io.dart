import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import 'token_storage_base.dart';

TokenStorage createTokenStorage() => SecureTokenStorage();

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
