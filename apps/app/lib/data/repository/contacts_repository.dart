import '../model/friendship.dart';
import '../remote/api_client.dart';

/// 好友域仓库（guide §3.4：widget 只依赖 repository）。
class ContactsRepository {
  const ContactsRepository(this._api);

  final ApiClient _api;

  /// 发起好友申请（防重复矩阵在服务端，业务错误经 ApiException 透传）。
  Future<void> sendRequest(int userId) async {
    await _api.post('/friends/requests', body: {'user_id': userId});
  }

  /// 我收到的待处理申请（新的在前）。
  Future<List<FriendRequestItem>> listRequests() async {
    final items = await _api.getList('/friends/requests');
    return items
        .map((e) => FriendRequestItem.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// 同意申请。
  Future<void> accept(int requestId) async {
    await _api.post('/friends/requests/$requestId/accept');
  }

  /// 拒绝申请。
  Future<void> reject(int requestId) async {
    await _api.post('/friends/requests/$requestId/reject');
  }

  /// 我的好友列表（结交时间升序）。
  Future<List<FriendItem>> listFriends() async {
    final items = await _api.getList('/friends');
    return items
        .map((e) => FriendItem.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
