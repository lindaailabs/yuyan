/// 用户资料（GET/PUT /users/me）。
class UserProfile {
  const UserProfile({
    required this.id,
    required this.phone,
    required this.avatarId,
    required this.createdAt,
    this.nickname,
  });

  final int id;
  final String phone;
  final String? nickname; // null = 未完成首登引导

  /// 预置头像 id（1~8）。
  final int avatarId;

  /// Unix 秒。
  final int createdAt;

  factory UserProfile.fromJson(Map<String, dynamic> json) => UserProfile(
        id: json['id'] as int,
        phone: json['phone'] as String,
        nickname: json['nickname'] as String?,
        avatarId: json['avatar_id'] as int,
        createdAt: json['created_at'] as int,
      );

  UserProfile copyWith({String? nickname, int? avatarId}) => UserProfile(
        id: id,
        phone: phone,
        nickname: nickname ?? this.nickname,
        avatarId: avatarId ?? this.avatarId,
        createdAt: createdAt,
      );
}

/// 双 token（登录/刷新响应）。
class TokenPair {
  const TokenPair({
    required this.accessToken,
    required this.refreshToken,
    required this.expiresIn,
  });

  final String accessToken;
  final String refreshToken;

  /// access 有效期（秒）。
  final int expiresIn;

  factory TokenPair.fromJson(Map<String, dynamic> json) => TokenPair(
        accessToken: json['access_token'] as String,
        refreshToken: json['refresh_token'] as String,
        expiresIn: json['expires_in'] as int,
      );
}

/// 验证码下发响应：一期为图形验证码（base64 PNG）；生产接短信后 image 为 null。
class SmsCodeResponse {
  const SmsCodeResponse({this.captchaImage});

  final String? captchaImage;

  factory SmsCodeResponse.fromJson(Map<String, dynamic> json) =>
      SmsCodeResponse(captchaImage: json['captcha_image'] as String?);
}

/// 搜索结果项（手机号脱敏）。
class UserSearchItem {
  const UserSearchItem({
    required this.id,
    required this.phone,
    required this.avatarId,
    this.nickname,
  });

  final int id;
  final String? nickname;
  final int avatarId;

  /// 形如 138****8000。
  final String phone;

  factory UserSearchItem.fromJson(Map<String, dynamic> json) => UserSearchItem(
        id: json['id'] as int,
        nickname: json['nickname'] as String?,
        avatarId: json['avatar_id'] as int,
        phone: json['phone'] as String,
      );
}
