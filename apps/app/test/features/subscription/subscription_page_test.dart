import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/l10n/zh.dart';
import 'package:yuyan_app/core/providers.dart';
import 'package:yuyan_app/data/remote/api_exception.dart';
import 'package:yuyan_app/features/subscription/subscription_page.dart';

import '../../support/fake_api_client.dart';

Map<String, dynamic> _entJson(String plan) => {
      'user_id': 10,
      'plan': plan,
      'status': 1,
      'quota': {
        'daily_messages': plan == 'pro' ? 500 : 50,
        'daily_used': 1,
        'daily_remain': plan == 'pro' ? 499 : 49,
        'memory_limit': plan == 'pro' ? 1000 : 200,
        'advanced_model': plan == 'pro',
      },
    };

void main() {
  late FakeApiClient api;

  setUp(() => api = FakeApiClient());

  Widget buildPage() => ProviderScope(
        overrides: [apiClientProvider.overrideWithValue(api)],
        child: const MaterialApp(home: SubscriptionPage()),
      );

  Future<void> pumpFrames(WidgetTester tester) async {
    for (var i = 0; i < 3; i++) {
      await tester.pump(const Duration(milliseconds: 200));
    }
  }

  testWidgets('免费版：展示套餐与额度', (tester) async {
    api.handlers['GET /entitlements/me'] = () async => _entJson('free');
    api.handlers['POST /events'] = () async => {'accepted': 1};

    await tester.pumpWidget(buildPage());
    await pumpFrames(tester);

    expect(find.text(Zh.subscription), findsOneWidget);
    expect(find.text(Zh.planFree), findsOneWidget);
    expect(find.text(Zh.subscriptionUpgradePro), findsOneWidget);
  });

  testWidgets('错误态：展示服务端文案', (tester) async {
    api.handlers['GET /entitlements/me'] =
        () async => throw const ApiException(1002, '未登录');

    await tester.pumpWidget(buildPage());
    await pumpFrames(tester);

    expect(find.text('未登录'), findsOneWidget);
  });
}
