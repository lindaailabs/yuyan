/// 权益与额度视图模型（服务端为唯一事实源，详见 add-entitlement-analytics 设计）。
library;

class QuotaView {
  const QuotaView({
    required this.dailyMessages,
    required this.dailyUsed,
    required this.dailyRemain,
    required this.memoryLimit,
    required this.advancedModel,
  });

  final int dailyMessages;
  final int dailyUsed;
  final int dailyRemain;
  final int memoryLimit;
  final bool advancedModel;

  factory QuotaView.fromJson(Map<String, dynamic> json) => QuotaView(
        dailyMessages: json['daily_messages'] as int? ?? 0,
        dailyUsed: json['daily_used'] as int? ?? 0,
        dailyRemain: json['daily_remain'] as int? ?? 0,
        memoryLimit: json['memory_limit'] as int? ?? 0,
        advancedModel: json['advanced_model'] as bool? ?? false,
      );
}

class EntitlementView {
  const EntitlementView({
    required this.userId,
    required this.plan,
    required this.status,
    required this.quota,
    this.renewAt,
    required this.updatedAt,
  });

  final int userId;
  final String plan;
  final int status;
  final QuotaView quota;
  final int? renewAt;
  final int updatedAt;

  factory EntitlementView.fromJson(Map<String, dynamic> json) => EntitlementView(
        userId: json['user_id'] as int? ?? 0,
        plan: (json['plan'] as String?) ?? Plan.free,
        status: json['status'] as int? ?? 1,
        quota: QuotaView.fromJson(json['quota'] as Map<String, dynamic>? ?? {}),
        renewAt: json['renew_at'] as int?,
        updatedAt: json['updated_at'] as int? ?? 0,
      );

  bool get isPro => plan == Plan.pro;
}

class Plan {
  const Plan._(this.value);

  final String value;

  static const free = 'free';
  static const pro = 'pro';

  static const List<String> values = [free, pro];
}
