import 'pet_memory.dart';

/// 消息角色（服务端 role 字段）。
enum PetMessageRole {
  user,
  assistant;

  static PetMessageRole parse(String? v) =>
      v == 'assistant' ? PetMessageRole.assistant : PetMessageRole.user;

  String get value => name;
}

/// 客户端发送状态（本地视图：乐观插入 → 成功/失败）。
enum PetSendState { sending, sent, failed }

/// 一条对话消息（服务端 pet_messages 的客户端表示）。
class PetMessage {
  const PetMessage({
    required this.id,
    required this.convId,
    required this.role,
    required this.content,
    required this.createdAt,
    this.clientMsgId,
    this.status = 1,
    this.errorCode = 0,
    this.sendState = PetSendState.sent,
  });

  /// 服务端消息 id；本地尚未落库（发送中）为 0。
  final int id;

  final int convId;
  final PetMessageRole role;
  final String content;

  /// 幂等键（UUID），失败重试复用同一个。
  final String? clientMsgId;

  /// 服务端状态：1=正常 2=失败。
  final int status;

  final int errorCode;
  final PetSendState sendState;
  final int createdAt;

  bool get isUser => role == PetMessageRole.user;
  bool get isFailed => sendState == PetSendState.failed || status == 2;

  factory PetMessage.fromJson(
    Map<String, dynamic> json, {
    PetSendState sendState = PetSendState.sent,
  }) =>
      PetMessage(
        id: json['id'] as int? ?? 0,
        convId: json['conv_id'] as int? ?? 0,
        role: PetMessageRole.parse(json['role'] as String?),
        content: json['content'] as String? ?? '',
        clientMsgId: json['client_msg_id'] as String?,
        status: json['status'] as int? ?? 1,
        errorCode: json['error_code'] as int? ?? 0,
        sendState: sendState,
        createdAt: json['created_at'] as int? ?? 0,
      );

  PetMessage copyWith({
    int? id,
    int? convId,
    String? content,
    PetSendState? sendState,
    int? status,
    int? errorCode,
  }) =>
      PetMessage(
        id: id ?? this.id,
        convId: convId ?? this.convId,
        role: role,
        content: content ?? this.content,
        clientMsgId: clientMsgId,
        status: status ?? this.status,
        errorCode: errorCode ?? this.errorCode,
        sendState: sendState ?? this.sendState,
        createdAt: createdAt,
      );
}

/// 历史消息游标分页结果（禁止 offset，guide §12）。
class MessagePage {
  const MessagePage({
    required this.items,
    required this.nextCursor,
    required this.hasMore,
  });

  final List<PetMessage> items;
  final int nextCursor;
  final bool hasMore;

  factory MessagePage.fromJson(Map<String, dynamic> json) => MessagePage(
        items: (json['items'] as List<dynamic>? ?? const [])
            .map((e) => PetMessage.fromJson(e as Map<String, dynamic>))
            .toList(),
        nextCursor: json['next_cursor'] as int? ?? 0,
        hasMore: json['has_more'] as bool? ?? false,
      );
}

/// 单次 AI 调用用量（guide §5：模型/token/耗时/错误码）。
class AiUsage {
  const AiUsage({
    this.provider = '',
    this.model = '',
    this.promptVersion = '',
    this.inputTokens = 0,
    this.outputTokens = 0,
    this.latencyMs = 0,
    this.cacheHit = false,
    this.errCode = 0,
  });

  final String provider;
  final String model;
  final String promptVersion;
  final int inputTokens;
  final int outputTokens;
  final int latencyMs;
  final bool cacheHit;
  final int errCode;

  factory AiUsage.fromJson(Map<String, dynamic> json) => AiUsage(
        provider: json['provider'] as String? ?? '',
        model: json['model'] as String? ?? '',
        promptVersion: json['prompt_version'] as String? ?? '',
        inputTokens: json['input_tokens'] as int? ?? 0,
        outputTokens: json['output_tokens'] as int? ?? 0,
        latencyMs: json['latency_ms'] as int? ?? 0,
        cacheHit: json['cache_hit'] as bool? ?? false,
        errCode: json['err_code'] as int? ?? 0,
      );
}

/// 发送结果（POST /pet-messages 的 data）。
class SendMessageResult {
  const SendMessageResult({
    required this.conversationId,
    required this.userMessage,
    this.assistantMessage,
    this.newMemories = const [],
    this.streaming = false,
    this.usage = const AiUsage(),
  });

  final int conversationId;
  final PetMessage userMessage;

  /// 为空表示回复尚未生成（幂等重放），UI 按待回复处理。
  final PetMessage? assistantMessage;

  /// 本轮新形成的长期记忆（聊天页轻量提示）。
  final List<PetMemory> newMemories;
  final bool streaming;
  final AiUsage usage;

  factory SendMessageResult.fromJson(Map<String, dynamic> json) =>
      SendMessageResult(
        conversationId: json['conversation_id'] as int? ?? 0,
        userMessage: PetMessage.fromJson(
          json['user_message'] as Map<String, dynamic>? ?? const {},
        ),
        assistantMessage: json['assistant_message'] == null
            ? null
            : PetMessage.fromJson(
                json['assistant_message'] as Map<String, dynamic>,
              ),
        newMemories: (json['new_memories'] as List<dynamic>? ?? const [])
            .map((e) => PetMemory.fromJson(e as Map<String, dynamic>))
            .toList(),
        streaming: json['streaming'] as bool? ?? false,
        usage: AiUsage.fromJson(
          json['usage'] as Map<String, dynamic>? ?? const {},
        ),
      );
}

/// 会话（POST /pet-conversations 的 data）。
class PetConversation {
  const PetConversation({
    required this.id,
    required this.userId,
    required this.petId,
    this.lastMsgPreview = '',
  });

  final int id;
  final int userId;
  final int petId;
  final String lastMsgPreview;

  factory PetConversation.fromJson(Map<String, dynamic> json) =>
      PetConversation(
        id: json['id'] as int? ?? 0,
        userId: json['user_id'] as int? ?? 0,
        petId: json['pet_id'] as int? ?? 0,
        lastMsgPreview: json['last_msg_preview'] as String? ?? '',
      );
}
