import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'core/router/app_router.dart';
import 'core/theme/app_theme.dart';
import 'data/local/app_database.dart';

void main() {
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

/// 数据库 provider（drift 单例；应用退出时关闭）。
final appDatabaseProvider = Provider<AppDatabase>((ref) {
  final db = AppDatabase();
  ref.onDispose(db.close);
  return db;
});

/// 路由 provider（便于测试注入）。
final appRouterProvider = Provider((ref) => appRouter);
