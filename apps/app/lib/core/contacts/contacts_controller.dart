import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/l10n/zh.dart';
import '../../data/model/friendship.dart';
import '../../data/repository/contacts_repository.dart';
import '../../data/remote/api_exception.dart';

/// 好友域状态：好友列表 + 待处理申请 + 各操作进度。
class ContactsState {
  const ContactsState({
    this.friends = const [],
    this.requests = const [],
    this.loading = false,
    this.processingIds = const {},
    this.error,
  });

  final List<FriendItem> friends;
  final List<FriendRequestItem> requests;

  /// 列表首载/刷新中。
  final bool loading;

  /// 正在 accept/reject 的申请 id（按钮防抖）。
  final Set<int> processingIds;

  /// 最近一次操作/加载失败信息（页面 snackbar 呈现）。
  final String? error;

  ContactsState copyWith({
    List<FriendItem>? friends,
    List<FriendRequestItem>? requests,
    bool? loading,
    Set<int>? processingIds,
    String? error,
    bool clearError = false,
  }) {
    return ContactsState(
      friends: friends ?? this.friends,
      requests: requests ?? this.requests,
      loading: loading ?? this.loading,
      processingIds: processingIds ?? this.processingIds,
      error: clearError ? null : (error ?? this.error),
    );
  }
}

/// 好友域状态机：列表加载 + 同意/拒绝局部更新（成功即改内存，页面免整刷）。
class ContactsController extends StateNotifier<ContactsState> {
  ContactsController(this._repo) : super(const ContactsState());

  final ContactsRepository _repo;

  /// 加载好友 + 申请双列表（页面进入时调用，幂等）。
  Future<void> loadAll() async {
    state = state.copyWith(loading: true, clearError: true);
    try {
      final results = await Future.wait<Object>([
        _repo.listFriends(),
        _repo.listRequests(),
      ]);
      state = ContactsState(
        friends: results[0] as List<FriendItem>,
        requests: results[1] as List<FriendRequestItem>,
      );
    } on ApiException catch (e) {
      state = state.copyWith(loading: false, error: e.msg);
    } catch (_) {
      state = state.copyWith(loading: false, error: Zh.errorOccurred);
    }
  }

  /// 同意申请：成功后从申请列表移除并追加到好友列表（新好友结交时间最晚，排最后）。
  Future<bool> accept(int requestId) async {
    final processing = {...state.processingIds}..add(requestId);
    state = state.copyWith(processingIds: processing, clearError: true);
    try {
      await _repo.accept(requestId);
    } on ApiException catch (e) {
      _finishProcessing(requestId, error: e.msg);
      return false;
    } catch (_) {
      _finishProcessing(requestId, error: Zh.errorOccurred);
      return false;
    }

    FriendRequestItem? matched;
    final requests = <FriendRequestItem>[];
    for (final r in state.requests) {
      if (r.id == requestId) {
        matched = r;
      } else {
        requests.add(r);
      }
    }
    final friends = [...state.friends];
    if (matched != null) {
      friends.add(FriendItem(user: matched.fromUser, createdAt: _nowSec()));
    }
    _finishProcessing(requestId);
    state = state.copyWith(requests: requests, friends: friends);
    return true;
  }

  /// 拒绝申请：仅从申请列表移除，好友列表不变。
  Future<bool> reject(int requestId) async {
    final processing = {...state.processingIds}..add(requestId);
    state = state.copyWith(processingIds: processing, clearError: true);
    try {
      await _repo.reject(requestId);
    } on ApiException catch (e) {
      _finishProcessing(requestId, error: e.msg);
      return false;
    } catch (_) {
      _finishProcessing(requestId, error: Zh.errorOccurred);
      return false;
    }

    final requests =
        state.requests.where((r) => r.id != requestId).toList();
    _finishProcessing(requestId);
    state = state.copyWith(requests: requests);
    return true;
  }

  /// 发起申请（搜索页调用）：错误原样上抛由 UI 决定呈现。
  Future<void> sendRequest(int userId) => _repo.sendRequest(userId);

  void _finishProcessing(int requestId, {String? error}) {
    final processing = {...state.processingIds}..remove(requestId);
    state = state.copyWith(processingIds: processing, error: error);
  }

  int _nowSec() => DateTime.now().millisecondsSinceEpoch ~/ 1000;
}
