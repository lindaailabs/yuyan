import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/auth/auth_state.dart';
import '../../core/l10n/zh.dart';
import '../../core/providers.dart';
import '../../data/remote/api_exception.dart';
import '../shared/avatar_widget.dart';

/// 资料页：展示 + 编辑昵称/头像，登出入口。
class ProfilePage extends ConsumerStatefulWidget {
  const ProfilePage({super.key});

  @override
  ConsumerState<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends ConsumerState<ProfilePage> {
  final _nicknameCtrl = TextEditingController();
  bool _editing = false;
  bool _loading = false;
  int? _pendingAvatarId;

  @override
  void dispose() {
    _nicknameCtrl.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final nickname = _nicknameCtrl.text.trim();
    if (nickname.isEmpty || nickname.characters.length > 20) {
      _showError(Zh.onboardingBadNickname);
      return;
    }
    setState(() => _loading = true);
    try {
      await ref.read(authControllerProvider.notifier).updateProfile(
            nickname: nickname,
            avatarId: _pendingAvatarId,
          );
      if (mounted) setState(() => _editing = false);
    } on ApiException catch (e) {
      _showError(e.msg);
    } catch (_) {
      _showError(Zh.errorOccurred);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _logout() async {
    await ref.read(authControllerProvider.notifier).logout();
    // 状态 → LoggedOut，路由守卫自动跳回 /login。
  }

  void _showError(String msg) {
    if (!mounted) return;
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(authControllerProvider);
    if (state is! AuthReady) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    final profile = state.profile;
    if (!_editing && _nicknameCtrl.text != (profile.nickname ?? '')) {
      _nicknameCtrl.text = profile.nickname ?? '';
    }

    return Scaffold(
      appBar: AppBar(
        title: Text(_editing ? Zh.profileEdit : Zh.profileTitle),
        actions: [
          TextButton(
            onPressed: _loading
                ? null
                : () => setState(() {
                      _editing = !_editing;
                      _pendingAvatarId = null;
                      _nicknameCtrl.text = profile.nickname ?? '';
                    }),
            child: Text(_editing ? Zh.cancel : Zh.profileEdit),
          ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(24),
              children: [
                Center(
                  child: _editing
                      ? _avatarPicker(profile.avatarId)
                      : AvatarWidget(avatarId: profile.avatarId, size: 96),
                ),
                const SizedBox(height: 24),
                ListTile(
                  leading: const Icon(Icons.phone),
                  title: Text(profile.phone),
                ),
                const SizedBox(height: 8),
                _editing
                    ? TextField(
                        controller: _nicknameCtrl,
                        maxLength: 20,
                        decoration: InputDecoration(
                          labelText: Zh.profileNicknameLabel,
                          counterText: '',
                          border: const OutlineInputBorder(),
                        ),
                      )
                    : ListTile(
                        leading: const Icon(Icons.badge),
                        title: Text(profile.nickname ?? ''),
                      ),
                const SizedBox(height: 32),
                if (_editing)
                  FilledButton(
                    onPressed: _save,
                    child: const Text(Zh.profileSave),
                  ),
                if (!_editing)
                  OutlinedButton(
                    onPressed: _logout,
                    child: const Text(Zh.profileLogout),
                  ),
              ],
            ),
    );
  }

  Widget _avatarPicker(int current) {
    final selected = _pendingAvatarId ?? current;
    return SizedBox(
      height: 96,
      child: ListView(
        scrollDirection: Axis.horizontal,
        children: [
          for (final id in List.generate(8, (i) => i + 1))
            Padding(
              padding: const EdgeInsets.only(right: 8),
              child: GestureDetector(
                onTap: () => setState(() => _pendingAvatarId = id),
                child: Container(
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    border: Border.all(
                      color: id == selected
                          ? Theme.of(context).colorScheme.primary
                          : Colors.transparent,
                      width: 3,
                    ),
                  ),
                  child: AvatarWidget(avatarId: id),
                ),
              ),
            ),
        ],
      ),
    );
  }
}
