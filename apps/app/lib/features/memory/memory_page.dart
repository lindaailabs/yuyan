import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/l10n/zh.dart';
import '../../core/providers.dart';
import '../../data/model/pet_memory.dart';

const _bg = Color(0xFFFFF9F5);
const _primary = Color(0xFF7C6BF5);
const _primarySoft = Color(0xFF9C8CFF);
const _ink = Color(0xFF2B2B34);
const _muted = Color(0xFF7A7A88);
const _danger = Color(0xFFE5484D);

/// 记忆页：查看宠物记住的事，并可删除（删除后不再召回）。
class MemoryPage extends ConsumerStatefulWidget {
  const MemoryPage({super.key, required this.petId});

  final int petId;

  @override
  ConsumerState<MemoryPage> createState() => _MemoryPageState();
}

class _MemoryPageState extends ConsumerState<MemoryPage> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(memoryControllerProvider(widget.petId).notifier).load();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(memoryControllerProvider(widget.petId));

    return Scaffold(
      backgroundColor: _bg,
      appBar: AppBar(
        backgroundColor: _bg,
        surfaceTintColor: Colors.transparent,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new_rounded),
          onPressed: () => Navigator.of(context).maybePop(),
        ),
        title: const Text(
          Zh.memoryTitle,
          style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
        ),
      ),
      body: state.loading
          ? const Center(child: CircularProgressIndicator())
          : state.error != null && state.memories.isEmpty
              ? _ErrorView(
                  message: state.error!,
                  onRetry: () =>
                      ref.read(memoryControllerProvider(widget.petId).notifier).load(),
                )
              : state.memories.isEmpty
                  ? const _EmptyView()
                  : RefreshIndicator(
                      onRefresh: () => ref
                          .read(memoryControllerProvider(widget.petId).notifier)
                          .load(),
                      child: ListView.separated(
                        physics: const AlwaysScrollableScrollPhysics(),
                        padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
                        itemCount: state.memories.length,
                        separatorBuilder: (_, _) => const SizedBox(height: 12),
                        itemBuilder: (context, index) => _MemoryCard(
                          memory: state.memories[index],
                          deleting: state.deletingIds.contains(
                            state.memories[index].id,
                          ),
                          onDelete: () => _confirmDelete(state.memories[index]),
                        ),
                      ),
                    ),
    );
  }

  Future<void> _confirmDelete(PetMemory memory) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text(Zh.memoryDeleteTitle),
        content: Text(Zh.memoryDeleteBody),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text(Zh.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text(Zh.memoryDeleteConfirm),
          ),
        ],
      ),
    );
    if (ok != true || !mounted) return;
    final done = await ref
        .read(memoryControllerProvider(widget.petId).notifier)
        .delete(memory.id);
    if (!done && mounted) {
      final msg = ref.read(memoryControllerProvider(widget.petId)).error;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(msg ?? Zh.memoryDeleteFailed)),
      );
    }
  }
}

class _MemoryCard extends StatelessWidget {
  const _MemoryCard({
    required this.memory,
    required this.deleting,
    required this.onDelete,
  });

  final PetMemory memory;
  final bool deleting;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFEDE6FF)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.04),
            blurRadius: 12,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 14, 8, 14),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              width: 36,
              height: 36,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: const LinearGradient(
                  colors: [_primarySoft, _primary],
                ),
              ),
              child: const Icon(
                Icons.psychology_alt_rounded,
                size: 20,
                color: Colors.white,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    memory.content,
                    style: const TextStyle(
                      fontSize: 14,
                      height: 1.4,
                      color: _ink,
                    ),
                  ),
                  const SizedBox(height: 6),
                  Text(
                    '${Zh.memoryTypeLabel}：${memory.type} · ${Zh.memoryConfidence}：${memory.confidence.toStringAsFixed(2)}',
                    style: const TextStyle(fontSize: 11, color: _muted),
                  ),
                ],
              ),
            ),
            IconButton(
              onPressed: deleting ? null : onDelete,
              icon: deleting
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(
                      Icons.delete_outline_rounded,
                      color: _danger,
                      size: 20,
                    ),
              tooltip: Zh.memoryDeleteConfirm,
            ),
          ],
        ),
      ),
    );
  }
}

class _EmptyView extends StatelessWidget {
  const _EmptyView();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 88,
              height: 88,
              decoration: const BoxDecoration(
                shape: BoxShape.circle,
                gradient: LinearGradient(
                  colors: [Color(0x2E9C8CFF), Color(0x38FFB86B)],
                ),
              ),
              child: const Icon(
                Icons.psychology_alt_outlined,
                color: _primary,
                size: 40,
              ),
            ),
            const SizedBox(height: 20),
            const Text(
              Zh.memoryEmpty,
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: _ink,
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              Zh.memoryEmptyHint,
              textAlign: TextAlign.center,
              style: TextStyle(fontSize: 13, color: _muted),
            ),
          ],
        ),
      ),
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(message, textAlign: TextAlign.center),
            const SizedBox(height: 16),
            OutlinedButton.icon(
              onPressed: onRetry,
              icon: const Icon(Icons.refresh),
              label: const Text(Zh.retry),
            ),
          ],
        ),
      ),
    );
  }
}
