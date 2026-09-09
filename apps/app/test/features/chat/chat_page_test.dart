import 'package:drift/native.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/l10n/zh.dart';
import 'package:yuyan_app/core/providers.dart';
import 'package:yuyan_app/data/local/app_database.dart';
import 'package:yuyan_app/features/chat/chat_page.dart';

import '../../support/fake_api_client.dart';

Map<String, dynamic> _petJson() => {
      'id': 7,
      'user_id': 1,
      'name': '小燕',
      'species': 'swallow',
      'avatar_id': 1,
      'persona': null,
      'level': 2,
      'intimacy': 3,
      'mood': 'curious',
      'created_at': 1788900000,
      'updated_at': 1788900000,
    };

Map<String, dynamic> _msgJson(int id, String role, String content) => {
      'id': id,
      'conv_id': 1,
      'role': role,
      'content': content,
      'client_msg_id': null,
      'status': 1,
      'created_at': 1788900000 + id,
    };

void main() {
  late FakeApiClient api;
  late AppDatabase db;

  setUp(() {
    api = FakeApiClient();
    db = AppDatabase.test(NativeDatabase.memory());
    api.handlers['GET /pets'] = () async => {
          'items': [_petJson()]
        };
    api.handlers['POST /pet-conversations'] = () async => {
          'id': 1,
          'user_id': 1,
          'pet_id': 7,
          'last_msg_preview': '',
        };
  });

  tearDown(() async {
    await db.close();
  });

  /// 页面含「呼吸」循环动画，pumpAndSettle 永不收敛：改为按帧推进。
  Future<void> pumpFrames(WidgetTester tester) async {
    for (var i = 0; i < 3; i++) {
      await tester.pump(const Duration(milliseconds: 200));
    }
  }

  Widget buildPage() => ProviderScope(
        overrides: [
          apiClientProvider.overrideWithValue(api),
          appDatabaseProvider.overrideWithValue(db),
        ],
        child: const MaterialApp(home: ChatPage(petId: 7)),
      );

  testWidgets('空态：展示招呼语与引导胶囊', (tester) async {
    api.handlers['GET /pet-messages'] =
        () async => {'items': [], 'next_cursor': 0, 'has_more': false};

    await tester.pumpWidget(buildPage());
    await pumpFrames(tester);

    expect(find.text(Zh.chatGreeting), findsOneWidget);
    expect(find.text(Zh.chatEmptyHint), findsOneWidget);
    expect(find.text(Zh.chatChip1), findsOneWidget);
    expect(find.text(Zh.chatInputHint), findsOneWidget);
  });

  testWidgets('有历史消息：渲染用户与宠物气泡', (tester) async {
    api.handlers['GET /pet-messages'] = () async => {
          'items': [
            _msgJson(1, 'user', '你好呀'),
            _msgJson(2, 'assistant', '我在这儿陪你'),
          ],
          'next_cursor': 2,
          'has_more': false,
        };

    await tester.pumpWidget(buildPage());
    await pumpFrames(tester);

    expect(find.text('你好呀'), findsOneWidget);
    expect(find.text('我在这儿陪你'), findsOneWidget);
    expect(find.text(Zh.chatGreeting), findsNothing);
  });

  testWidgets('输入内容后发送按钮可用，点击后清空输入', (tester) async {
    var posted = false;
    api.handlers['GET /pet-messages'] =
        () async => {'items': [], 'next_cursor': 0, 'has_more': false};
    api.handlers['POST /pet-messages'] = () async {
      posted = true;
      return {
        'conversation_id': 1,
        'user_message': _msgJson(3, 'user', '嗨')
          ..['client_msg_id'] = 'cid-1',
        'assistant_message': _msgJson(4, 'assistant', '你好'),
        'streaming': false,
        'usage': {
          'provider': 'mock',
          'model': 'mock-pet-1',
          'input_tokens': 1,
          'output_tokens': 1,
        },
      };
    };

    await tester.pumpWidget(buildPage());
    await pumpFrames(tester);

    await tester.enterText(find.byType(TextField), '嗨');
    await pumpFrames(tester);
    await tester.tap(find.byIcon(Icons.send_rounded));
    await pumpFrames(tester);

    expect(posted, isTrue);
    expect(find.text('嗨'), findsWidgets);
  });
}
