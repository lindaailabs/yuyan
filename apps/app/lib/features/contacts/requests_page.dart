import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/contacts/contacts_controller.dart';
import '../../core/l10n/zh.dart';
import '../../core/providers.dart';
import '../../data/model/friendship.dart';
import '../shared/avatar_widget.dart';

/// 申请列表页：申请卡片 + 同意/拒绝（防抖禁用、失败 snackbar，tasks 3.4）。
class RequestsPage extends ConsumerWidget {
  const RequestsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    // accept/reject 失败 → snackbar（错误写入 state.error）。
    ref.listen<ContactsState>(contactsControllerProvider, (prev, next) {
      final msg = next.error;
      if (msg != null && msg != prev?.error) {
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(SnackBar(content: Text(msg)));
      }
    });

    final state = ref.watch(contactsControllerProvider);

    return Scaffold(
      appBar: AppBar(title: const Text(Zh.requestsTitle)),
      body: state.loading
          ? const Center(child: CircularProgressIndicator())
          : state.requests.isEmpty
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.mark_email_unread_outlined,
                          size: 72, color: Colors.grey),
                      const SizedBox(height: 16),
                      Text(Zh.requestsEmpty),
                    ],
                  ),
                )
              : ListView.separated(
                  itemCount: state.requests.length,
                  separatorBuilder: (_, _) => const Divider(height: 1),
                  itemBuilder: (context, index) {
                    final request = state.requests[index];
                    return _RequestCard(request: request);
                  },
                ),
    );
  }
}

/// 单条申请卡片：申请人资料 + 同意/拒绝按钮。
class _RequestCard extends ConsumerWidget {
  const _RequestCard({required this.request});

  final FriendRequestItem request;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(contactsControllerProvider);
    final busy = state.processingIds.contains(request.id);

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Row(
          children: [
            AvatarWidget(avatarId: request.fromUser.avatarId),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    request.fromUser.nickname ?? request.fromUser.phone,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 4),
                  Text(
                    request.fromUser.phone,
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ],
              ),
            ),
            const SizedBox(width: 8),
            FilledButton.tonal(
              onPressed: busy
                  ? null
                  : () => ref
                      .read(contactsControllerProvider.notifier)
                      .accept(request.id),
              child: busy
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text(Zh.requestsAccept),
            ),
            const SizedBox(width: 8),
            OutlinedButton(
              onPressed: busy
                  ? null
                  : () => ref
                      .read(contactsControllerProvider.notifier)
                      .reject(request.id),
              child: Text(Zh.requestsReject),
            ),
          ],
        ),
      ),
    );
  }
}
