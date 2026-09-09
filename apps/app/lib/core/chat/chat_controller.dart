import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/model/pet_message.dart';
import '../../data/remote/api_exception.dart';
import '../../data/repository/pet_chat_repository.dart';
import '../l10n/zh.dart';
import 'client_msg_id.dart';

/// 聊天页状态：本地优先渲染 + 增量同步 + 发送状态机。
class ChatState {
  const ChatState({
    this.messages = const [],
    this.loading = false,
    this.sending = false,
    this.syncing = false,
    this.hasMore = false,
    this.error,
  });

  final List<PetMessage> messages;

  /// 首次加载中。
  final bool loading;

  /// 正在发送一条消息（输入框防抖）。
  final bool sending;

  /// 增量/分页拉取中。
  final bool syncing;

  /// 服务端还有更多历史（可继续加载）。
  final bool hasMore;

  final String? error;

  ChatState copyWith({
    List<PetMessage>? messages,
    bool? loading,
    bool? sending,
    bool? syncing,
    bool? hasMore,
    String? error,
    bool clearError = false,
  }) =>
      ChatState(
        messages: messages ?? this.messages,
        loading: loading ?? this.loading,
        sending: sending ?? this.sending,
        syncing: syncing ?? this.syncing,
        hasMore: hasMore ?? this.hasMore,
        error: clearError ? null : error ?? this.error,
      );
}

/// 聊天状态机：打开会话（本地 → 增量）、发送（乐观插入/失败重试）、继续加载。
class ChatController extends StateNotifier<ChatState> {
  ChatController(this._repo, {required this.petId})
      : super(const ChatState());

  final PetChatRepository _repo;
  final int petId;

  /// 单次同步最多翻页数（避免超长会话一次拉爆）。
  static const int maxPagesPerSync = 5;

  int? _convId;
  bool _opened = false;

  int? get conversationId => _convId;

  /// 进入页面：先渲染本地缓存，再从服务端增量补齐。
  Future<void> open() async {
    if (_opened) return;
    _opened = true;
    state = state.copyWith(loading: true, clearError: true);
    try {
      _convId = await _repo.ensureConversation(petId);
      state = state.copyWith(
        messages: await _repo.localMessages(_convId!),
      );
      await sync();
      state = state.copyWith(loading: false);
    } on ApiException catch (e) {
      state = state.copyWith(loading: false, error: e.msg);
    } catch (_) {
      state = state.copyWith(loading: false, error: Zh.errorOccurred);
    }
  }

  /// 按游标从服务端增量补齐（本地无数据时即为全量回补）。
  Future<void> sync() async {
    final convId = _convId;
    if (convId == null) return;
    state = state.copyWith(syncing: true);
    try {
      var cursor = await _repo.cursorOf(convId);
      var hasMore = true;
      for (var page = 0; page < maxPagesPerSync && hasMore; page++) {
        final data = await _repo.fetchHistory(convId, cursor);
        await _repo.cacheServerMessages(data.items);
        cursor = data.nextCursor;
        hasMore = data.hasMore;
      }
      state = state.copyWith(
        messages: await _repo.localMessages(convId),
        hasMore: hasMore,
        syncing: false,
      );
    } on ApiException catch (e) {
      state = state.copyWith(syncing: false, error: e.msg);
    } catch (_) {
      state = state.copyWith(syncing: false, error: Zh.errorOccurred);
    }
  }

  /// 发送一条消息：乐观插入「发送中」→ 成功/失败回填。
  Future<bool> send(String text) async {
    final content = text.trim();
    if (content.isEmpty || state.sending) return false;
    return _deliver(content: content, clientMsgId: generateClientMsgId());
  }

  /// 重试失败消息：复用同一 client_msg_id，服务端幂等。
  Future<bool> retry(String clientMsgId) async {
    if (state.sending) return false;
    final target = state.messages
        .where((m) => m.clientMsgId == clientMsgId)
        .firstOrNull;
    if (target == null) return false;
    return _deliver(content: target.content, clientMsgId: clientMsgId);
  }

  Future<bool> _deliver({
    required String content,
    required String clientMsgId,
  }) async {
    final convId = _convId ?? await _ensureConversation();
    if (convId == null) return false;

    final optimistic = PetMessage(
      id: 0,
      convId: convId,
      role: PetMessageRole.user,
      content: content,
      clientMsgId: clientMsgId,
      sendState: PetSendState.sending,
      createdAt: _nowSec(),
    );
    await _repo.markSending(clientMsgId);
    if (!state.messages.any((m) => m.clientMsgId == clientMsgId)) {
      await _repo.cacheMessage(optimistic);
    }
    state = state.copyWith(
      messages: _upsert(optimistic),
      sending: true,
      clearError: true,
    );

    try {
      final res = await _repo.send(
        petId: petId,
        content: content,
        clientMsgId: clientMsgId,
      );
      _convId ??= res.conversationId;
      await _repo.markSent(clientMsgId, res.userMessage.id);
      if (res.assistantMessage != null) {
        await _repo.cacheMessage(res.assistantMessage!);
      }
      state = state.copyWith(
        messages: await _repo.localMessages(convId),
        sending: false,
      );
      return true;
    } on ApiException catch (e) {
      await _repo.markFailed(clientMsgId, e.code);
      state = state.copyWith(
        messages: await _repo.localMessages(convId),
        sending: false,
        error: e.msg,
      );
      return false;
    } catch (_) {
      await _repo.markFailed(clientMsgId, 0);
      state = state.copyWith(
        messages: await _repo.localMessages(convId),
        sending: false,
        error: Zh.chatSendFailed,
      );
      return false;
    }
  }

  Future<int?> _ensureConversation() async {
    try {
      final id = await _repo.ensureConversation(petId);
      if (id == 0) return null;
      _convId = id;
      return id;
    } on ApiException catch (e) {
      state = state.copyWith(error: e.msg);
      return null;
    } catch (_) {
      state = state.copyWith(error: Zh.errorOccurred);
      return null;
    }
  }

  /// 按 client_msg_id 插入或替换本地视图中的消息。
  List<PetMessage> _upsert(PetMessage m) {
    final next = <PetMessage>[];
    var replaced = false;
    for (final item in state.messages) {
      if (item.clientMsgId != null && item.clientMsgId == m.clientMsgId) {
        next.add(m);
        replaced = true;
      } else {
        next.add(item);
      }
    }
    if (!replaced) next.add(m);
    return next;
  }

  int _nowSec() => DateTime.now().millisecondsSinceEpoch ~/ 1000;
}
