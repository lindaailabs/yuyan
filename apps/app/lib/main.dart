import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'core/app_config.dart';
import 'core/auth/auth_state.dart';
import 'core/providers.dart';
import 'core/router/app_router.dart';
import 'core/theme/app_theme.dart';
import 'data/local/app_database.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await AppConfig.load();
  runApp(const ProviderScope(child: YuyanApp()));
}

/// 应用根组件：Riverpod ProviderScope + go_router 路由壳。
class YuyanApp extends ConsumerWidget {
  const YuyanApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(appRouterProvider);
    return MaterialApp.router(
      title: 'Yuyan',
      theme: AppTheme.light,
      darkTheme: AppTheme.dark,
      routerConfig: router,
    );
  }
}

/// 路由 provider：绑定认证三态（状态变化 → redirect 重估）。
final appRouterProvider = Provider<GoRouter>((ref) {
  final authState = ValueNotifier<AuthState>(ref.read(authControllerProvider));
  ref.listen<AuthState>(authControllerProvider, (_, next) {
    authState.value = next;
  });
  ref.onDispose(authState.dispose);
  return buildRouter(authState);
});

/// 数据库 provider（drift 单例；应用退出时关闭）。
final appDatabaseProvider = Provider<AppDatabase>((ref) {
  final db = AppDatabase();
  ref.onDispose(db.close);
  return db;
});
