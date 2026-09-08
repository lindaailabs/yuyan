import 'package:go_router/go_router.dart';

import '../../features/auth/login_page.dart';
import '../../features/home/home_page.dart';

/// 应用路由（go_router）。
/// W1 为壳：登录占位 + 主页占位；W2 起按功能逐步替换。
final appRouter = GoRouter(
  initialLocation: '/login',
  routes: [
    GoRoute(
      path: '/login',
      builder: (context, state) => const LoginPage(),
    ),
    GoRoute(
      path: '/home',
      builder: (context, state) => const HomePage(),
    ),
  ],
);
