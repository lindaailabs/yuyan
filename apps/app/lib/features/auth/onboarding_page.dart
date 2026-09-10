import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/l10n/zh.dart';
import '../../core/providers.dart';
import '../../data/remote/api_exception.dart';
import '../shared/avatar_widget.dart';

/// 首登引导页：8 头像选择 + 昵称输入 → 提交进入主页。
class OnboardingPage extends ConsumerStatefulWidget {
  const OnboardingPage({super.key});

  @override
  ConsumerState<OnboardingPage> createState() => _OnboardingPageState();
}

class _OnboardingPageState extends ConsumerState<OnboardingPage> {
  final _nicknameCtrl = TextEditingController();
  int _avatarId = 1;
  bool _loading = false;

  @override
  void dispose() {
    _nicknameCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final nickname = _nicknameCtrl.text.trim();
    if (nickname.isEmpty || nickname.characters.length > 20) {
      _showError(Zh.onboardingBadNickname);
      return;
    }
    setState(() => _loading = true);
    try {
      await ref
          .read(authControllerProvider.notifier)
          .completeOnboarding(nickname, _avatarId);
      // 状态 → Ready，路由守卫自动跳转 /home。
    } on ApiException catch (e) {
      _showError(e.msg);
    } catch (_) {
      _showError(Zh.errorOccurred);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _showError(String msg) {
    if (!mounted) return;
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text(Zh.onboardingTitle)),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              Zh.onboardingPickAvatar,
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 12),
            GridView.count(
              crossAxisCount: 4,
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              mainAxisSpacing: 12,
              crossAxisSpacing: 12,
              children: [
                for (final id in List.generate(
                  UserAvatarWidget.count,
                  (i) => i + 1,
                ))
                  _avatarTile(id),
              ],
            ),
            const SizedBox(height: 24),
            TextField(
              controller: _nicknameCtrl,
              maxLength: 20,
              decoration: InputDecoration(
                labelText: Zh.onboardingNicknameHint,
                counterText: '',
                border: const OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 24),
            FilledButton(
              onPressed: _loading ? null : _submit,
              child: _loading
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text(Zh.onboardingSubmit),
            ),
          ],
        ),
      ),
    );
  }

  Widget _avatarTile(int id) {
    final selected = id == _avatarId;
    return GestureDetector(
      onTap: () => setState(() => _avatarId = id),
      child: Container(
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          border: Border.all(
            color: selected
                ? Theme.of(context).colorScheme.primary
                : Colors.transparent,
            width: 3,
          ),
        ),
        child: UserAvatarWidget(avatarId: id),
      ),
    );
  }
}
