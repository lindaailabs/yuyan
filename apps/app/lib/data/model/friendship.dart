import 'user_profile.dart';

/// 好友申请列表项（GET /friends/requests）。
class FriendRequestItem {
  const FriendRequestItem({
    required this.id,
    required this.fromUser,
    required this.createdAt,
  });

  /// 申请行 id（accept/reject 路径参数）。
  final int id;

  /// 申请人资料（手机号脱敏）。
  final UserSearchItem fromUser;

  /// Unix 秒。
  final int createdAt;

  factory FriendRequestItem.fromJson(Map<String, dynamic> json) =>
      FriendRequestItem(
        id: json['id'] as int,
        fromUser: UserSearchItem.fromJson(json['from_user'] as Map<String, dynamic>),
        createdAt: json['created_at'] as int,
      );
}

/// 好友列表项（GET /friends）。
class FriendItem {
  const FriendItem({required this.user, required this.createdAt});

  /// 好友资料（手机号脱敏）。
  final UserSearchItem user;

  /// 结交时间，Unix 秒。
  final int createdAt;

  factory FriendItem.fromJson(Map<String, dynamic> json) => FriendItem(
        user: UserSearchItem.fromJson(json['user'] as Map<String, dynamic>),
        createdAt: json['created_at'] as int,
      );
}
