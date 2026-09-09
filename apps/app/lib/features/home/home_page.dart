import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/l10n/zh.dart';
import '../../core/providers.dart';
import '../../data/model/pet.dart';
import '../shared/avatar_widget.dart';

/// AI 宠物主页：一期主线从 IM 会话切到宠物陪伴闭环。
class HomePage extends ConsumerWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(petControllerProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text(Zh.homeTitle),
        actions: [
          IconButton(
            icon: const Icon(Icons.person),
            tooltip: Zh.homeProfile,
            onPressed: () => context.go('/profile'),
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () => ref.read(petControllerProvider.notifier).load(),
        child: ListView(
          padding: const EdgeInsets.all(20),
          children: [
            if (state.loading)
              const Padding(
                padding: EdgeInsets.only(top: 80),
                child: Center(child: CircularProgressIndicator()),
              )
            else if (state.current == null)
              _EmptyPetView(
                error: state.error,
                onCreate: () => context.go('/pet/create'),
                onRetry: () => ref.read(petControllerProvider.notifier).load(),
              )
            else
              _PetHomeView(
                pet: state.current!,
                error: state.error,
                onCreate: () => context.go('/pet/create'),
                onChat: () => context.go('/chat/${state.current!.id}'),
                onMemory: () => context.go('/memories/${state.current!.id}'),
                onGrowth: () => context.go('/growth/${state.current!.id}'),
              ),
          ],
        ),
      ),
    );
  }
}

class _EmptyPetView extends StatelessWidget {
  const _EmptyPetView({
    required this.onCreate,
    required this.onRetry,
    this.error,
  });

  final VoidCallback onCreate;
  final VoidCallback onRetry;
  final String? error;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SizedBox(height: 48),
        const Icon(Icons.auto_awesome, size: 72),
        const SizedBox(height: 24),
        Text(
          Zh.homePetEmptyTitle,
          textAlign: TextAlign.center,
          style: Theme.of(context).textTheme.headlineSmall,
        ),
        const SizedBox(height: 12),
        Text(
          Zh.homePetEmptyBody,
          textAlign: TextAlign.center,
          style: Theme.of(context).textTheme.bodyLarge,
        ),
        if (error != null) ...[
          const SizedBox(height: 20),
          Text(error!, textAlign: TextAlign.center),
          TextButton.icon(
            onPressed: onRetry,
            icon: const Icon(Icons.refresh),
            label: const Text(Zh.retry),
          ),
        ],
        const SizedBox(height: 28),
        FilledButton.icon(
          onPressed: onCreate,
          icon: const Icon(Icons.add),
          label: const Text(Zh.homePetCreate),
        ),
      ],
    );
  }
}

class _PetHomeView extends StatelessWidget {
  const _PetHomeView({
    required this.pet,
    required this.onCreate,
    required this.onChat,
    required this.onMemory,
    required this.onGrowth,
    this.error,
  });

  final PetProfile pet;
  final VoidCallback onCreate;
  final VoidCallback onChat;
  final VoidCallback onMemory;
  final VoidCallback onGrowth;
  final String? error;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          children: [
            AvatarWidget(avatarId: pet.avatarId, size: 88),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(pet.name, style: theme.textTheme.headlineSmall),
                  const SizedBox(height: 6),
                  Text(pet.species, style: theme.textTheme.bodyMedium),
                ],
              ),
            ),
            IconButton(
              onPressed: onCreate,
              icon: const Icon(Icons.add),
              tooltip: Zh.homePetCreate,
            ),
          ],
        ),
        const SizedBox(height: 24),
        Wrap(
          spacing: 12,
          runSpacing: 12,
          children: [
            _MetricTile(label: Zh.homePetMood, value: pet.mood),
            _MetricTile(label: Zh.homePetLevel, value: '${pet.level}'),
            _MetricTile(label: Zh.homePetIntimacy, value: '${pet.intimacy}'),
          ],
        ),
        if (error != null) ...[
          const SizedBox(height: 16),
          Text(error!, textAlign: TextAlign.center),
        ],
        const SizedBox(height: 28),
        _ActionButton(
          icon: Icons.chat_bubble_outline,
          label: Zh.homePetChat,
          onTap: onChat,
        ),
        const SizedBox(height: 12),
        _ActionButton(
          icon: Icons.psychology_alt_outlined,
          label: Zh.homePetMemory,
          onTap: onMemory,
        ),
        const SizedBox(height: 12),
        _ActionButton(
          icon: Icons.trending_up,
          label: Zh.homePetGrowth,
          onTap: onGrowth,
        ),
      ],
    );
  }
}

class _MetricTile extends StatelessWidget {
  const _MetricTile({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 104,
      child: DecoratedBox(
        decoration: BoxDecoration(
          border: Border.all(
            color: Theme.of(context).colorScheme.outlineVariant,
          ),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label, style: Theme.of(context).textTheme.labelMedium),
              const SizedBox(height: 8),
              Text(value, style: Theme.of(context).textTheme.titleMedium),
            ],
          ),
        ),
      ),
    );
  }
}

class _ActionButton extends StatelessWidget {
  const _ActionButton({
    required this.icon,
    required this.label,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return OutlinedButton.icon(
      onPressed: onTap,
      icon: Icon(icon),
      label: Text(label),
    );
  }
}
