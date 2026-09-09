import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/l10n/zh.dart';
import 'package:yuyan_app/core/providers.dart';
import 'package:yuyan_app/features/home/home_page.dart';

import 'support/fake_api_client.dart';

Map<String, dynamic> _petJson() => {
  'id': 1,
  'user_id': 10,
  'name': '小燕',
  'species': 'swallow',
  'avatar_id': 2,
  'persona': null,
  'level': 1,
  'intimacy': 0,
  'mood': 'curious',
  'created_at': 1788858000,
  'updated_at': 1788858000,
};

void main() {
  testWidgets('HomePage 无宠物时展示领养入口', (tester) async {
    final api = FakeApiClient();
    api.handlers['GET /pets'] = () async => {'items': []};

    await tester.pumpWidget(
      ProviderScope(
        overrides: [apiClientProvider.overrideWithValue(api)],
        child: const MaterialApp(home: HomePage()),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text(Zh.homeTitle), findsOneWidget);
    expect(find.text(Zh.homePetEmptyTitle), findsOneWidget);
    expect(find.text(Zh.homePetCreate), findsOneWidget);
    expect(find.byTooltip(Zh.homeProfile), findsOneWidget);
    expect(find.byTooltip(Zh.homeContacts), findsNothing);
    expect(find.byTooltip(Zh.homeRequests), findsNothing);
  });

  testWidgets('HomePage 有宠物时展示状态与后续能力入口', (tester) async {
    final api = FakeApiClient();
    api.handlers['GET /pets'] = () async => {
      'items': [_petJson()],
    };

    await tester.pumpWidget(
      ProviderScope(
        overrides: [apiClientProvider.overrideWithValue(api)],
        child: const MaterialApp(home: HomePage()),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('小燕'), findsOneWidget);
    expect(find.text(Zh.homePetMood), findsOneWidget);
    expect(find.text(Zh.homePetLevel), findsOneWidget);
    expect(find.text(Zh.homePetIntimacy), findsOneWidget);
    expect(find.text(Zh.homePetChat), findsOneWidget);
    expect(find.text(Zh.homePetMemory), findsOneWidget);
    expect(find.text(Zh.homePetGrowth), findsOneWidget);
  });
}
