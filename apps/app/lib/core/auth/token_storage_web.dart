import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:web/web.dart' as web;

import 'token_storage_base.dart';

TokenStorage createTokenStorage() => WebTokenStorage();

/// Web token 存储。
///
/// flutter_secure_storage 的 WebCrypto 实现要求安全上下文。`localhost` 与
/// HTTPS 通常可用，但 `http://192.168.x.x` 本地联调不是安全上下文，会导致登录
/// 接口成功后 token 写入失败。这里在安全存储不可用时退到 sessionStorage，只保留
/// 当前浏览器会话，避免影响移动端生产实现。
class WebTokenStorage implements TokenStorage {
  WebTokenStorage({FlutterSecureStorage? secureStorage})
    : _secureStorage = secureStorage ?? const FlutterSecureStorage();

  static const _keyAccess = 'auth.access_token';
  static const _keyRefresh = 'auth.refresh_token';

  final FlutterSecureStorage _secureStorage;
  bool _useSessionFallback = false;

  @override
  Future<String?> readAccess() => _read(_keyAccess);

  @override
  Future<String?> readRefresh() => _read(_keyRefresh);

  @override
  Future<void> write({required String access, required String refresh}) async {
    if (!_useSessionFallback) {
      try {
        await _secureStorage.write(key: _keyAccess, value: access);
        await _secureStorage.write(key: _keyRefresh, value: refresh);
        return;
      } catch (_) {
        _useSessionFallback = true;
      }
    }
    web.window.sessionStorage.setItem(_keyAccess, access);
    web.window.sessionStorage.setItem(_keyRefresh, refresh);
  }

  @override
  Future<void> clear() async {
    if (!_useSessionFallback) {
      try {
        await _secureStorage.delete(key: _keyAccess);
        await _secureStorage.delete(key: _keyRefresh);
      } catch (_) {
        _useSessionFallback = true;
      }
    }
    web.window.sessionStorage.removeItem(_keyAccess);
    web.window.sessionStorage.removeItem(_keyRefresh);
  }

  Future<String?> _read(String key) async {
    if (!_useSessionFallback) {
      try {
        final value = await _secureStorage.read(key: key);
        if (value != null) {
          return value;
        }
      } catch (_) {
        _useSessionFallback = true;
      }
    }
    return web.window.sessionStorage.getItem(key);
  }
}
