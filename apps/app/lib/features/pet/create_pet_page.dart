import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/l10n/zh.dart';
import '../../core/providers.dart';
import '../shared/avatar_widget.dart';

class CreatePetPage extends ConsumerStatefulWidget {
  const CreatePetPage({super.key});

  @override
  ConsumerState<CreatePetPage> createState() => _CreatePetPageState();
}

class _CreatePetPageState extends ConsumerState<CreatePetPage> {
  final _nameController = TextEditingController();
  int _avatarId = 1;
  String? _nameError;

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(petControllerProvider);

    return Scaffold(
      appBar: AppBar(title: const Text(Zh.petCreateTitle)),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          TextField(
            controller: _nameController,
            maxLength: 20,
            decoration: InputDecoration(
              labelText: Zh.petNameHint,
              errorText: _nameError,
            ),
          ),
          const SizedBox(height: 16),
          Text(
            Zh.petPickAvatar,
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 12),
          Wrap(
            spacing: 12,
            runSpacing: 12,
            children: [
              for (var i = 1; i <= PetAvatarWidget.count; i++)
                ChoiceChip(
                  selected: _avatarId == i,
                  label: PetAvatarWidget(avatarId: i, size: 44),
                  onSelected: (_) => setState(() => _avatarId = i),
                ),
            ],
          ),
          if (state.error != null) ...[
            const SizedBox(height: 16),
            Text(state.error!, textAlign: TextAlign.center),
          ],
          const SizedBox(height: 28),
          FilledButton.icon(
            onPressed: state.creating ? null : _submit,
            icon: state.creating
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.favorite_border),
            label: const Text(Zh.petCreateSubmit),
          ),
        ],
      ),
    );
  }

  Future<void> _submit() async {
    final name = _nameController.text.trim();
    if (name.isEmpty || name.runes.length > 20) {
      setState(() => _nameError = Zh.petBadName);
      return;
    }
    setState(() => _nameError = null);
    final ok = await ref
        .read(petControllerProvider.notifier)
        .create(name: name, avatarId: _avatarId);
    if (!mounted) {
      return;
    }
    if (ok) {
      context.go('/home');
    }
  }
}
