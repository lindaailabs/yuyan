/// AI 宠物档案与状态模型。
class PetProfile {
  const PetProfile({
    required this.id,
    required this.userId,
    required this.name,
    required this.species,
    required this.avatarId,
    required this.level,
    required this.intimacy,
    required this.mood,
    required this.createdAt,
    required this.updatedAt,
    this.persona,
  });

  final int id;
  final int userId;
  final String name;
  final String species;
  final int avatarId;
  final String? persona;
  final int level;
  final int intimacy;
  final String mood;
  final int createdAt;
  final int updatedAt;

  factory PetProfile.fromJson(Map<String, dynamic> json) => PetProfile(
    id: json['id'] as int,
    userId: json['user_id'] as int,
    name: json['name'] as String,
    species: json['species'] as String,
    avatarId: json['avatar_id'] as int,
    persona: json['persona'] as String?,
    level: json['level'] as int,
    intimacy: json['intimacy'] as int,
    mood: json['mood'] as String,
    createdAt: json['created_at'] as int,
    updatedAt: json['updated_at'] as int,
  );
}

class PetStateView {
  const PetStateView({
    required this.id,
    required this.name,
    required this.species,
    required this.avatarId,
    required this.level,
    required this.intimacy,
    required this.mood,
  });

  final int id;
  final String name;
  final String species;
  final int avatarId;
  final int level;
  final int intimacy;
  final String mood;

  factory PetStateView.fromJson(Map<String, dynamic> json) => PetStateView(
    id: json['id'] as int,
    name: json['name'] as String,
    species: json['species'] as String,
    avatarId: json['avatar_id'] as int,
    level: json['level'] as int,
    intimacy: json['intimacy'] as int,
    mood: json['mood'] as String,
  );
}
