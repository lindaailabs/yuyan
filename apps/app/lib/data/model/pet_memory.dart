/// 长期记忆（pet_memories 的客户端表示）。
class PetMemory {
  const PetMemory({
    required this.id,
    required this.petId,
    required this.type,
    required this.content,
    required this.confidence,
    required this.createdAt,
    this.lastUsedAt,
    this.sourceMsgId,
  });

  final int id;
  final int petId;

  /// 记忆类型：preference / profile / event / relation。
  final String type;

  final String content;

  /// 置信度 0~1。
  final double confidence;
  final int createdAt;
  final int? lastUsedAt;
  final int? sourceMsgId;

  factory PetMemory.fromJson(Map<String, dynamic> json) => PetMemory(
        id: json['id'] as int? ?? 0,
        petId: json['pet_id'] as int? ?? 0,
        type: json['type'] as String? ?? '',
        content: json['content'] as String? ?? '',
        confidence: (json['confidence'] as num?)?.toDouble() ?? 0,
        createdAt: json['created_at'] as int? ?? 0,
        lastUsedAt: json['last_used_at'] as int?,
        sourceMsgId: json['source_msg_id'] as int?,
      );
}

/// 记忆类型的中文标签（文案收敛在 l10n，此处仅做类型→键的映射标识）。
String memoryTypeLabel(String type) => type;
