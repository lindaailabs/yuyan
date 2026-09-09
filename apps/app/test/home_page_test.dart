import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/l10n/zh.dart';
import 'package:yuyan_app/core/providers.dart';
import 'package:yuyan_app/features/home/home_page.dart';

import 'support/fake_api_client.dart';

void main() {
  testWidgets('HomePage 渲染 l10n 文案占位与通讯录/申请入口', (tester) async {
    final api = FakeApiClient();
    api.handlers['GET /friends'] = () async => {'items': []};
    api.handlers['GET /friends/requests'] = () async => {'items': []};

    await tester.pumpWidget(
      ProviderScope(
        overrides: [apiClientProvider.overrideWithValue(api)],
        child: const MaterialApp(home: HomePage()),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text(Zh.homeTitle), findsOneWidget);
    expect(find.text(Zh.homePlaceholder), findsOneWidget);
    // W3 新增入口：通讯录 + 好友申请。
    expect(find.byTooltip(Zh.homeContacts), findsOneWidget);
    expect(find.byTooltip(Zh.homeRequests), findsOneWidget);
  });

  testWidgets('有未处理申请时主页申请入口显示数量角标', (tester) async {
    final api = FakeApiClient();
    api.handlers['GET /friends'] = () async => {'items': []};
    api.handlers['GET /friends/requests'] = () async => {
          'items': [
            {
              'id': 10,
              'from_user': {
                'id': 4,
                'phone': '138****0004',
                'nickname': '用户4',
                'avatar_id': 4,
              },
              'created_at': 1788858000,
            }
          ]
        };

    await tester.pumpWidget(
      ProviderScope(
        overrides: [apiClientProvider.overrideWithValue(api)],
        child: const MaterialApp(home: HomePage()),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.byType(Badge), findsOneWidget);
    expect(find.text('1'), findsOneWidget);
  });
}
