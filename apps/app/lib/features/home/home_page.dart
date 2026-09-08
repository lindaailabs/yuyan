import 'package:flutter/material.dart';

import '../../core/l10n/zh.dart';

/// 主页（W1 占位；W5 实现会话列表）。
class HomePage extends StatelessWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text(Zh.homeTitle)),
      body: const Center(child: Text(Zh.homePlaceholder)),
    );
  }
}
