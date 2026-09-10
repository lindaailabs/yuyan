/// Token 安全存储抽象。
abstract class TokenStorage {
  Future<String?> readAccess();
  Future<String?> readRefresh();

  /// 写入双 token。
  Future<void> write({required String access, required String refresh});

  /// 清空（登出/会话失效）。
  Future<void> clear();
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
