import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/l10n/zh.dart';
import '../../core/providers.dart';
import '../shared/avatar_widget.dart';

/// 通讯录页：好友卡片列表 + 空态引导 + 失败重试（tasks 3.3）。
class ContactsPage extends ConsumerWidget {
  const ContactsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(contactsControllerProvider);

    return Scaffold(
      appBar: AppBar(title: const Text(Zh.contactsTitle)),
      body: state.loading
          ? const Center(child: CircularProgressIndicator())
          : state.error != null && state.friends.isEmpty
          ? _ErrorView(
              message: state.error!,
              onRetry: () =>
                  ref.read(contactsControllerProvider.notifier).loadAll(),
            )
          : state.friends.isEmpty
          ? _EmptyView(onGoAdd: () => context.go('/search'))
          : RefreshIndicator(
              onRefresh: () =>
                  ref.read(contactsControllerProvider.notifier).loadAll(),
              child: ListView.separated(
                physics: const AlwaysScrollableScrollPhysics(),
                itemCount: state.friends.length,
                separatorBuilder: (_, _) => const Divider(height: 1),
                itemBuilder: (context, index) {
                  final friend = state.friends[index];
                  return ListTile(
                    leading: UserAvatarWidget(avatarId: friend.user.avatarId),
                    title: Text(friend.user.nickname ?? friend.user.phone),
                    subtitle: Text(friend.user.phone),
                  );
                },
              ),
            ),
    );
  }
}

/// 空态：引导去搜索页添加好友。
class _EmptyView extends StatelessWidget {
  const _EmptyView({required this.onGoAdd});

  final VoidCallback onGoAdd;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.people_outline, size: 72, color: Colors.grey),
          const SizedBox(height: 16),
          Text(Zh.contactsEmpty, style: Theme.of(context).textTheme.bodyMedium),
          const SizedBox(height: 24),
          FilledButton.tonal(onPressed: onGoAdd, child: Text(Zh.contactsGoAdd)),
        ],
      ),
    );
  }
}

/// 失败态：错误信息 + 重试。
class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(message, textAlign: TextAlign.center),
          const SizedBox(height: 16),
          OutlinedButton(onPressed: onRetry, child: Text(Zh.retry)),
        ],
      ),
    );
  }
}
