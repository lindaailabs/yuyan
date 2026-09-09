import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/l10n/zh.dart';
import '../../core/providers.dart';
import '../../data/remote/api_exception.dart';

/// 注册页：手机号 + 密码 + 确认密码，注册成功即登录。
class RegisterPage extends ConsumerStatefulWidget {
  const RegisterPage({super.key});

  @override
  ConsumerState<RegisterPage> createState() => _RegisterPageState();
}

class _RegisterPageState extends ConsumerState<RegisterPage> {
  final _phoneCtrl = TextEditingController();
  final _pwdCtrl = TextEditingController();
  final _confirmCtrl = TextEditingController();
  bool _loading = false;

  @override
  void dispose() {
    _phoneCtrl.dispose();
    _pwdCtrl.dispose();
    _confirmCtrl.dispose();
    super.dispose();
  }

  bool _isValidPhone(String phone) => RegExp(r'^1[3-9]\d{9}$').hasMatch(phone);

  Future<void> _submit() async {
    final phone = _phoneCtrl.text.trim();
    final password = _pwdCtrl.text;
    final confirm = _confirmCtrl.text;
    if (!_isValidPhone(phone)) {
      _showError(Zh.registerBadPhone);
      return;
    }
    if (password.length < 6 || password.length > 64) {
      _showError(Zh.registerBadPassword);
      return;
    }
    if (password != confirm) {
      _showError(Zh.registerPasswordMismatch);
      return;
    }
    setState(() => _loading = true);
    try {
      await ref.read(authControllerProvider.notifier).register(phone, password);
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
      appBar: AppBar(title: const Text(Zh.registerTitle)),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            TextField(
              controller: _phoneCtrl,
              keyboardType: TextInputType.phone,
              maxLength: 11,
              decoration: InputDecoration(
                labelText: Zh.registerPhoneHint,
                counterText: '',
                border: const OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _pwdCtrl,
              obscureText: true,
              decoration: InputDecoration(
                labelText: Zh.registerPasswordHint,
                border: const OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _confirmCtrl,
              obscureText: true,
              decoration: InputDecoration(
                labelText: Zh.registerConfirmHint,
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
                  : Text(Zh.registerSubmit),
            ),
            const SizedBox(height: 12),
            Center(
              child: TextButton(
                onPressed: _loading ? null : () => context.pop(),
                child: Text(Zh.registerToLogin),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
