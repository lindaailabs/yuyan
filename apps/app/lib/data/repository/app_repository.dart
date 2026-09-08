import '../remote/api_client.dart';
import '../remote/ws_client.dart';

/// 跨 remote/local 的仓库抽象占位（guide §3.4：remote+local 合并策略）。
/// W2 起按 feature 拆分 auth/contacts/chat/conversation 仓库。
abstract class AppRepository {
  ApiClient get api;
  WsClient get ws;
}
