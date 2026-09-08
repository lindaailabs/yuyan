/// 远程网络层：WebSocket 客户端抽象（guide §3.4：widget 树禁止直接调用）。
/// W1 为占位接口；W4 实现连接管理、心跳、JSON 帧编解码（协议见 packages/protocol）。
abstract class WsClient {
  Future<void> connect(Uri uri);
  Future<void> send(String rawFrame);
  Stream<String> get frames;
  Future<void> close();
}
