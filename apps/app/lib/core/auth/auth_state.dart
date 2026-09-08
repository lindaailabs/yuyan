import '../../data/model/user_profile.dart';

/// 认证三态（guide §5.3 首登判定 + MILESTONES W2）：
/// - LoggedOut：未登录 → 路由守卫定向 /login
/// - Onboarding：已登录但 nickname 为 null（未完成首登引导）→ /onboarding
/// - Ready：登录且资料完备 → /home
sealed class AuthState {
  const AuthState();
}

/// 未登录。
class AuthLoggedOut extends AuthState {
  const AuthLoggedOut();
}

/// 已登录、待引导（提交昵称+头像后进入 Ready）。
class AuthOnboarding extends AuthState {
  const AuthOnboarding();
}

/// 登录就绪。
class AuthReady extends AuthState {
  const AuthReady({required this.profile});

  final UserProfile profile;
}
