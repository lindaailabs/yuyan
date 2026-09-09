/// 成长事件（pet_growth_events 的客户端表示）。
class GrowthEvent {
  const GrowthEvent({
    required this.id,
    required this.petId,
    required this.eventType,
    required this.reason,
    required this.createdAt,
    this.delta,
    this.sourceMsgId,
  });

  final int id;
  final int petId;

  /// message / level_up / mood_change / streak_milestone。
  final String eventType;

  /// 可解释的原因（服务端生成）。
  final String reason;

  /// 增量 JSON 字符串（如 {"intimacy":2}），可能为 null。
  final String? delta;
  final int? sourceMsgId;
  final int createdAt;

  factory GrowthEvent.fromJson(Map<String, dynamic> json) => GrowthEvent(
        id: json['id'] as int? ?? 0,
        petId: json['pet_id'] as int? ?? 0,
        eventType: json['event_type'] as String? ?? '',
        reason: json['reason'] as String? ?? '',
        delta: json['delta'] as String?,
        sourceMsgId: json['source_msg_id'] as int?,
        createdAt: json['created_at'] as int? ?? 0,
      );
}

/// 事件类型的展示文案键（文案本体在 l10n）。
String growthEventTypeKey(String type) => switch (type) {
      'level_up' => 'growthTypeLevelUp',
      'mood_change' => 'growthTypeMood',
      'streak_milestone' => 'growthTypeStreak',
      _ => 'growthTypeMessage',
    };
