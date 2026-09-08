import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../core/l10n/zh.dart';

/// 登录页（W1 占位；W2 实现手机号+图形验证码登录）。
class LoginPage extends StatelessWidget {
  const LoginPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text(Zh.loginTitle)),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(Zh.loginPhoneHint),
            const SizedBox(height: 8),
            Text(Zh.loginCodeHint),
            const SizedBox(height: 24),
            FilledButton(
              onPressed: () => context.go('/home'),
              child: const Text(Zh.loginSubmit),
            ),
          ],
        ),
      ),
    );
  }
}
