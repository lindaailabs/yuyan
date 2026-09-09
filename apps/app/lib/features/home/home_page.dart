import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/l10n/zh.dart';
import '../../core/providers.dart';

/// 主页（W2：会话占位 + 资料/搜索/通讯录/申请入口；W5 替换为会话列表）。
class HomePage extends ConsumerWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final requests = ref.watch(
        contactsControllerProvider.select((s) => s.requests.length));

    return Scaffold(
      appBar: AppBar(
        title: const Text(Zh.homeTitle),
        actions: [
          IconButton(
            icon: const Icon(Icons.contacts),
            tooltip: Zh.homeContacts,
            onPressed: () => context.go('/contacts'),
          ),
          Badge.count(
            count: requests,
            isLabelVisible: requests > 0,
            child: IconButton(
              icon: const Icon(Icons.person_add_alt),
              tooltip: Zh.homeRequests,
              onPressed: () => context.go('/requests'),
            ),
          ),
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
