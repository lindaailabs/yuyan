import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../core/l10n/zh.dart';

/// 主页（W2：会话占位 + 资料/搜索入口；W5 替换为会话列表）。
class HomePage extends StatelessWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text(Zh.homeTitle),
        actions: [
          IconButton(
            icon: const Icon(Icons.person),
            tooltip: Zh.homeProfile,
            onPressed: () => context.go('/profile'),
          ),
          IconButton(
            icon: const Icon(Icons.person_search),
            tooltip: Zh.homeSearch,
            onPressed: () => context.go('/search'),
          ),
        ],
      ),
      body: const Center(child: Text(Zh.homePlaceholder)),
    );
  }
}
