import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/l10n/zh.dart';
import 'package:yuyan_app/core/providers.dart';
import 'package:yuyan_app/features/growth/growth_page.dart';

import '../../support/fake_api_client.dart';

Map<String, dynamic> _eventJson(int id, String type, String reason) => {
      'id': id,
      'pet_id': 7,
      'event_type': type,
      'reason': reason,
      'delta': '{"intimacy":2}',
      'created_at': 1788900000 + id,
    };

void main() {
  late FakeApiClient api;

  setUp(() {
    api = FakeApiClient();
  });

  Widget buildPage() => ProviderScope(
        overrides: [apiClientProvider.overrideWithValue(api)],
        child: const MaterialApp(home: GrowthPage(petId: 7)),
      );

  Future<void> pumpFrames(WidgetTester tester) async {
    for (var i = 0; i < 3; i++) {
      await tester.pump(const Duration(milliseconds: 200));
    }
  }

  testWidgets('空态：无成长事件时展示引导', (tester) async {
    api.handlers['GET /pets/7/growth-events'] = () async => {'items': []};

    await tester.pumpWidget(buildPage());
    await pumpFrames(tester);

    expect(find.text(Zh.growthTitle), findsOneWidget);
    expect(find.text(Zh.growthEmpty), findsOneWidget);
  });

  testWidgets('时间线：展示事件标题与原因', (tester) async {
    api.handlers['GET /pets/7/growth-events'] = () async => {
          'items': [
            _eventJson(2, 'level_up', '亲密度达到 20，从 1 级升到 2 级'),
            _eventJson(1, 'message', '完成一次对话，亲密度 +2'),
          ]
        };

    await tester.pumpWidget(buildPage());
    await pumpFrames(tester);

    expect(find.text(Zh.growthTypeLevelUp), findsOneWidget);
    expect(find.text(Zh.growthTypeMessage), findsOneWidget);
    expect(find.text('完成一次对话，亲密度 +2'), findsOneWidget);
  });
}
