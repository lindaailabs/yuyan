import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/model/user_profile.dart';
import '../../data/repository/auth_repository.dart';
import '../../data/remote/api_exception.dart';
import 'auth_state.dart';
import 'token_storage.dart';

/// 认证状态机：登录 → 引导 → 就绪；会话失效回落未登录。
class AuthController extends StateNotifier<AuthState> {
  AuthController(this._repo, this._storage) : super(const AuthLoggedOut());

  final AuthRepository _repo;
  final TokenStorage _storage;

  /// 启动恢复：本地有 token → 拉取 me 判定三态。
  Future<void> init() async {
    final access = await _storage.readAccess();
    if (access == null || access.isEmpty) {
      state = const AuthLoggedOut();
      return;
    }
    try {
      final profile = await _repo.getMe();
      state = profile.nickname == null
          ? const AuthOnboarding()
          : AuthReady(profile: profile);
    } on ApiException {
      // 会话失效（401 且刷新也失败）或资料异常：清库回落未登录，用户重新登录。
      await _storage.clear();
      state = const AuthLoggedOut();
    }
  }

  /// 手机号+密码登录：成功后按 nickname 分流 Onboarding/Ready。
  Future<void> login(String phone, String password) async {
    final pair = await _repo.login(phone, password);
    await _storage.write(access: pair.accessToken, refresh: pair.refreshToken);
    final profile = await _repo.getMe();
    state = profile.nickname == null
        ? const AuthOnboarding()
        : AuthReady(profile: profile);
  }

  /// 手机号+密码注册（自动登录）：成功后按 nickname 分流 Onboarding/Ready。
  Future<void> register(String phone, String password) async {
    final pair = await _repo.register(phone, password);
    await _storage.write(access: pair.accessToken, refresh: pair.refreshToken);
    final profile = await _repo.getMe();
    state = profile.nickname == null
        ? const AuthOnboarding()
        : AuthReady(profile: profile);
  }

  /// 首登引导提交：昵称 + 头像 → Ready。
  Future<void> completeOnboarding(String nickname, int avatarId) async {
    final profile = await _repo.updateMe(nickname: nickname, avatarId: avatarId);
    state = AuthReady(profile: profile);
  }

  /// 资料页更新（昵称/头像单独或同时）。
  Future<UserProfile> updateProfile({String? nickname, int? avatarId}) async {
    final profile = await _repo.updateMe(nickname: nickname, avatarId: avatarId);
    state = AuthReady(profile: profile);
    return profile;
  }

  /// 拦截器会话失效回调：清库回落未登录。
  Future<void> forceLogout() async {
    await _storage.clear();
    state = const AuthLoggedOut();
  }

  /// 用户主动登出。
  Future<void> logout() => forceLogout();
}
