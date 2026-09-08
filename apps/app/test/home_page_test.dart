import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/l10n/zh.dart';
import 'package:yuyan_app/features/home/home_page.dart';

void main() {
  testWidgets('HomePage 渲染 l10n 文案占位', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: HomePage()),
    );
    expect(find.text(Zh.homeTitle), findsOneWidget);
    expect(find.text(Zh.homePlaceholder), findsOneWidget);
  });
}
