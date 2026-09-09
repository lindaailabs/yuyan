import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/l10n/zh.dart';
import '../../core/providers.dart';
import '../../data/model/pet_growth.dart';

const _bg = Color(0xFFFFF9F5);
const _primary = Color(0xFF7C6BF5);
const _accent = Color(0xFFFFB86B);
const _ink = Color(0xFF2B2B34);
const _muted = Color(0xFF7A7A88);

/// 成长时间线页：宠物每一次可解释的变化。
class GrowthPage extends ConsumerStatefulWidget {
  const GrowthPage({super.key, required this.petId});

  final int petId;

  @override
  ConsumerState<GrowthPage> createState() => _GrowthPageState();
}

class _GrowthPageState extends ConsumerState<GrowthPage> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(growthControllerProvider(widget.petId).notifier).load();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(growthControllerProvider(widget.petId));

    return Scaffold(
      backgroundColor: _bg,
      appBar: AppBar(
        backgroundColor: _bg,
        surfaceTintColor: Colors.transparent,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new_rounded),
          onPressed: () {
            // 主页钻取进入时（push）直接 pop 回主页；
            // 若从 URL 直接打开（无历史栈）则回主页兜底。
            if (Navigator.of(context).canPop()) {
              Navigator.of(context).pop();
            } else {
              context.go('/home');
            }
          },
        ),
        title: const Text(
          Zh.growthTitle,
          style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
        ),
      ),
      body: state.loading
          ? const Center(child: CircularProgressIndicator())
          : state.error != null && state.events.isEmpty
              ? _ErrorView(
                  message: state.error!,
                  onRetry: () =>
                      ref.read(growthControllerProvider(widget.petId).notifier).load(),
                )
              : state.events.isEmpty
                  ? const _EmptyView()
                  : RefreshIndicator(
                      onRefresh: () => ref
                          .read(growthControllerProvider(widget.petId).notifier)
                          .load(),
                      child: ListView.builder(
                        physics: const AlwaysScrollableScrollPhysics(),
                        padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
                        itemCount: state.events.length,
                        itemBuilder: (context, index) => _TimelineTile(
                          event: state.events[index],
                          isLast: index == state.events.length - 1,
                        ),
                      ),
                    ),
    );
  }
}

class _TimelineTile extends StatelessWidget {
  const _TimelineTile({required this.event, required this.isLast});

  final GrowthEvent event;
  final bool isLast;

  @override
  Widget build(BuildContext context) {
    final color = switch (event.eventType) {
      'level_up' => _accent,
      'mood_change' => const Color(0xFF2ECC8F),
      'streak_milestone' => _primary,
      _ => const Color(0xFF9C8CFF),
    };
    return IntrinsicHeight(
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          SizedBox(
            width: 24,
            child: Column(
              children: [
                Container(
                  width: 12,
                  height: 12,
                  decoration: BoxDecoration(
                    color: color,
                    shape: BoxShape.circle,
                  ),
                ),
                if (!isLast)
                  Expanded(
                    child: Container(width: 2, color: const Color(0xFFEDE6FF)),
                  ),
              ],
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.only(bottom: 18),
              child: DecoratedBox(
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: const Color(0xFFEDE6FF)),
                ),
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        _titleOf(event),
                        style: const TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: _ink,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        event.reason,
                        style: const TextStyle(fontSize: 12, color: _muted),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  String _titleOf(GrowthEvent e) => switch (e.eventType) {
        'level_up' => Zh.growthTypeLevelUp,
        'mood_change' => Zh.growthTypeMood,
        'streak_milestone' => Zh.growthTypeStreak,
        _ => Zh.growthTypeMessage,
      };
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
                  colors: [Color(0x2EFFB86B), Color(0x2E9C8CFF)],
                ),
              ),
              child: const Icon(Icons.trending_up, color: _primary, size: 40),
            ),
            const SizedBox(height: 20),
            const Text(
              Zh.growthEmpty,
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: _ink,
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              Zh.growthEmptyHint,
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
