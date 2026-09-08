import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/l10n/zh.dart';
import '../../core/providers.dart';
import '../../data/remote/api_exception.dart';

/// 登录页：手机号 + 验证码。
/// 图形验证码模式：展示 base64 图片，点击刷新；
/// 生产短信模式（captcha_image 缺省）：提示等待短信。
class LoginPage extends ConsumerStatefulWidget {
  const LoginPage({super.key});

  @override
  ConsumerState<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends ConsumerState<LoginPage> {
  final _phoneCtrl = TextEditingController();
  final _codeCtrl = TextEditingController();

  /// base64 图形验证码（null = 短信模式提示态）。
  String? _captchaImage;
  bool _loading = false;
  String? _hint;

  @override
  void dispose() {
    _phoneCtrl.dispose();
    _codeCtrl.dispose();
    super.dispose();
  }

  Future<void> _requestCode() async {
    final phone = _phoneCtrl.text.trim();
    if (!_isValidPhone(phone)) {
      _showError(Zh.loginBadPhone);
      return;
    }
    setState(() => _loading = true);
    try {
      final resp = await ref.read(authRepositoryProvider).sendSmsCode(phone);
      setState(() {
        _captchaImage = resp.captchaImage;
        _hint = resp.captchaImage == null ? Zh.loginSmsSent : null;
      });
    } on ApiException catch (e) {
      _showError(e.msg);
    } catch (_) {
      _showError(Zh.errorOccurred);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _submit() async {
    final phone = _phoneCtrl.text.trim();
    final code = _codeCtrl.text.trim();
    if (!_isValidPhone(phone)) {
      _showError(Zh.loginBadPhone);
      return;
    }
    if (code.isEmpty) {
      _showError(Zh.loginEmptyCode);
      return;
    }
    setState(() => _loading = true);
    try {
      // 登录成功后 authControllerProvider 状态变化 → 路由守卫自动跳转。
      await ref.read(authControllerProvider.notifier).login(phone, code);
    } on ApiException catch (e) {
      _showError(e.msg);
    } catch (_) {
      _showError(Zh.errorOccurred);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  bool _isValidPhone(String phone) => RegExp(r'^1[3-9]\d{9}$').hasMatch(phone);

  void _showError(String msg) {
    if (!mounted) return;
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text(Zh.loginTitle)),
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
                labelText: Zh.loginPhoneHint,
                counterText: '',
                border: const OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 16),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: TextField(
                    controller: _codeCtrl,
                    keyboardType: TextInputType.number,
                    maxLength: 6,
                    decoration: InputDecoration(
                      labelText: Zh.loginCodeHint,
                      counterText: '',
                      border: const OutlineInputBorder(),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                OutlinedButton(
                  onPressed: _loading ? null : _requestCode,
                  child: Text(Zh.loginGetCode),
                ),
              ],
            ),
            const SizedBox(height: 16),
            if (_captchaImage != null)
              GestureDetector(
                onTap: _requestCode,
                child: Center(
                  child: Column(
                    children: [
                      Image.memory(
                        base64Decode(_captchaImage!),
                        height: 40,
                        fit: BoxFit.contain,
                      ),
                      const SizedBox(height: 4),
                      Text(
                        Zh.loginCaptchaRefresh,
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                  ),
                ),
              )
            else if (_hint != null)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(_hint!),
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
                  : Text(Zh.loginSubmit),
            ),
          ],
        ),
      ),
    );
  }
}
