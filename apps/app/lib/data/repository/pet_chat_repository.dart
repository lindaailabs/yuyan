import 'package:drift/drift.dart';

import '../local/app_database.dart';
import '../model/pet_message.dart';
import '../remote/api_client.dart';

/// 宠物对话仓库：远端 REST + 本地 drift 缓存的合并策略（guide §3.3：UI 不直接调 Dio/DAO）。
class PetChatRepository {
  const PetChatRepository(this._api, this._dao);

  final ApiClient _api;
  final MessageDao _dao;

  /// 创建或获取会话，返回会话 id。
  Future<int> ensureConversation(int petId) async {
    final data = await _api.post('/pet-conversations', body: {'pet_id': petId});
    return data['id'] as int? ?? 0;
  }

  /// 按游标拉取历史（id > cursor 升序，禁止 offset）。
  Future<MessagePage> fetchHistory(
    int convId,
    int cursor, {
    int limit = 20,
  }) async {
    final data = await _api.get(
      '/pet-messages',
      query: {
        'conv_id': '$convId',
        'cursor': '$cursor',
        'limit': '$limit',
      },
    );
    return MessagePage.fromJson(data);
  }

  /// 发送消息（携带 client_msg_id 保证重试幂等）。
  Future<SendMessageResult> send({
    required int petId,
    required String content,
    required String clientMsgId,
  }) async {
    final data = await _api.post(
      '/pet-messages',
      body: {
        'pet_id': petId,
        'content': content,
        'client_msg_id': clientMsgId,
      },
    );
    return SendMessageResult.fromJson(data);
  }

  /// 本地缓存的消息（进入页面先渲染，再增量同步）。
  Future<List<PetMessage>> localMessages(int convId) async {
    final rows = await _dao.messagesOf(convId);
    return rows.map(_fromRow).toList();
  }

  /// 本地已落库的最大服务端 id（增量拉取游标）。
  Future<int> cursorOf(int convId) => _dao.maxServerId(convId);

  /// 缓存服务端返回的消息：本地已存在的用户消息按 client_msg_id 回填，其余新增。
  Future<void> cacheServerMessages(List<PetMessage> msgs) async {
    for (final m in msgs) {
      var handled = false;
      if (m.isUser && m.clientMsgId != null) {
        handled = await markSent(m.clientMsgId!, m.id) > 0;
      }
      if (!handled) {
        await cacheMessage(m);
      }
    }
  }

  /// 写入本地（发送中的乐观消息或服务端已确认消息）。
  Future<void> cacheMessage(PetMessage m) => _dao.insertMessage(
        LocalMessagesCompanion.insert(
          convId: m.convId,
          serverId: m.id == 0 ? const Value.absent() : Value(m.id),
          role: m.role.value,
          content: m.content,
          clientMsgId: Value(m.clientMsgId),
          sendState: Value(_sendStateValue(m.sendState)),
          errorCode: Value(m.errorCode),
          createdAt: m.createdAt,
        ),
      );

  /// 发送成功：按 client_msg_id 回填服务端 id 与状态（不新增行），返回影响行数。
  Future<int> markSent(String clientMsgId, int serverId) =>
      _dao.updateByClientMsgId(
        clientMsgId,
        LocalMessagesCompanion(
          serverId: Value(serverId),
          sendState: const Value(LocalSendState.sent),
        ),
      );

  /// 发送失败：标记失败态与错误码，保留内容供重试。
  Future<void> markFailed(String clientMsgId, int errorCode) =>
      _dao.updateByClientMsgId(
        clientMsgId,
        LocalMessagesCompanion(
          sendState: const Value(LocalSendState.failed),
          errorCode: Value(errorCode),
        ),
      );

  /// 重试前置：把失败消息置回发送中。
  Future<void> markSending(String clientMsgId) => _dao.updateByClientMsgId(
        clientMsgId,
        const LocalMessagesCompanion(
          sendState: Value(LocalSendState.sending),
        ),
      );

  static int _sendStateValue(PetSendState s) => switch (s) {
        PetSendState.sending => LocalSendState.sending,
        PetSendState.sent => LocalSendState.sent,
        PetSendState.failed => LocalSendState.failed,
      };

  static PetSendState _sendStateOf(int v) => switch (v) {
        LocalSendState.sent => PetSendState.sent,
        LocalSendState.failed => PetSendState.failed,
        _ => PetSendState.sending,
      };

  static PetMessage _fromRow(LocalMessageRow row) => PetMessage(
        id: row.serverId ?? 0,
        convId: row.convId,
        role: PetMessageRole.parse(row.role),
        content: row.content,
        clientMsgId: row.clientMsgId,
        status: 1,
        errorCode: row.errorCode,
        sendState: _sendStateOf(row.sendState),
        createdAt: row.createdAt,
      );
}
