import 'dart:math';

/// 生成 36 字符的 client_msg_id（UUID 形制），用于发送幂等与失败重试。
/// 服务端对该字段校验 len=36；重复提交同一 id 不会产生重复消息。
String generateClientMsgId() {
  final rnd = Random();
  String hex(int len) => List.generate(
        len,
        (_) => rnd.nextInt(16).toRadixString(16),
      ).join();
  return '${hex(8)}-${hex(4)}-${hex(4)}-${hex(4)}-${hex(12)}';
}
