import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/l10n/zh.dart';
import 'package:yuyan_app/core/providers.dart';
import 'package:yuyan_app/features/memory/memory_page.dart';

import '../../support/fake_api_client.dart';

Map<String, dynamic> _memoryJson(int id, String content) => {
      'id': id,
      'pet_id': 7,
      'type': 'preference',
      'content': content,
      'confidence': 0.8,
      'created_at': 1788900000 + id,
    };

void main() {
  late FakeApiClient api;

  setUp(() {
    api = FakeApiClient();
  });

  Widget buildPage() => ProviderScope(
        overrides: [apiClientProvider.overrideWithValue(api)],
        child: const MaterialApp(home: MemoryPage(petId: 7)),
      );

  Future<void> pumpFrames(WidgetTester tester) async {
    for (var i = 0; i < 3; i++) {
      await tester.pump(const Duration(milliseconds: 200));
    }
  }

  testWidgets('空态：无记忆时展示引导', (tester) async {
    api.handlers['GET /pets/7/memories'] = () async => {'items': []};

    await tester.pumpWidget(buildPage());
    await pumpFrames(tester);

    expect(find.text(Zh.memoryTitle), findsOneWidget);
    expect(find.text(Zh.memoryEmpty), findsOneWidget);
    expect(find.text(Zh.memoryEmptyHint), findsOneWidget);
  });

  testWidgets('列表：展示记忆内容与删除入口', (tester) async {
    api.handlers['GET /pets/7/memories'] = () async => {
          'items': [_memoryJson(1, '用户喜欢蓝色')]
        };

    await tester.pumpWidget(buildPage());
    await pumpFrames(tester);

    expect(find.text('用户喜欢蓝色'), findsOneWidget);
    expect(find.byIcon(Icons.delete_outline_rounded), findsOneWidget);
  });

  testWidgets('删除：确认后调用接口并从列表移除', (tester) async {
    var deleted = false;
    api.handlers['GET /pets/7/memories'] = () async => {
          'items': [_memoryJson(1, '用户喜欢蓝色')]
        };
    api.handlers['DELETE /pet-memories/1'] = () async {
      deleted = true;
      return {};
    };

    await tester.pumpWidget(buildPage());
    await pumpFrames(tester);

    await tester.tap(find.byIcon(Icons.delete_outline_rounded));
    await tester.pumpAndSettle();
    expect(find.text(Zh.memoryDeleteTitle), findsOneWidget);

    await tester.tap(find.text(Zh.memoryDeleteConfirm));
    await pumpFrames(tester);

    expect(deleted, isTrue);
    expect(find.text('用户喜欢蓝色'), findsNothing);
    expect(find.text(Zh.memoryEmpty), findsOneWidget);
  });
}
