/// 服务端统一包裹的业务错误：code != 0（分段见 guide §5.3）。
class ApiException implements Exception {
  const ApiException(this.code, this.msg);

  final int code;
  final String msg;

  /// 会话失效（1002 / HTTP 401）。
  bool get isUnauthorized => code == 1002;

  @override
  String toString() => 'ApiException($code, $msg)';
}
