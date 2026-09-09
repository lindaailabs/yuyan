import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/chat/chat_controller.dart';
import 'package:yuyan_app/core/l10n/zh.dart';
import 'package:yuyan_app/data/local/app_database.dart';
import 'package:yuyan_app/data/model/pet_message.dart';
import 'package:yuyan_app/data/remote/api_exception.dart';
import 'package:yuyan_app/data/repository/pet_chat_repository.dart';

import '../../support/fake_api_client.dart';

Map<String, dynamic> _msgJson(
  int id,
  String role,
  String content, {
  String? clientMsgId,
}) =>
    {
      'id': id,
      'conv_id': 1,
      'role': role,
      'content': content,
      'client_msg_id': clientMsgId,
      'status': 1,
      'created_at': 1788900000 + id,
    };

Map<String, dynamic> _sendResult({
  required String clientMsgId,
  required String userContent,
}) =>
    {
      'conversation_id': 1,
      'user_message': _msgJson(3, 'user', userContent, clientMsgId: clientMsgId),
      'assistant_message': _msgJson(4, 'assistant', '我在这儿呢'),
      'streaming': false,
      'usage': {
        'provider': 'mock',
        'model': 'mock-pet-1',
        'prompt_version': 'pet-chat-v1',
        'input_tokens': 10,
        'output_tokens': 5,
        'latency_ms': 12,
        'cache_hit': false,
        'err_code': 0,
      },
    };

void main() {
  late FakeApiClient api;
  late AppDatabase db;
  late PetChatRepository repo;
  late ChatController controller;

  setUp(() {
    api = FakeApiClient();
    db = AppDatabase.test(NativeDatabase.memory());
    repo = PetChatRepository(api, MessageDao(db));
    controller = ChatController(repo, petId: 7);

    api.handlers['POST /pet-conversations'] =
        () async => {'id': 1, 'user_id': 1, 'pet_id': 7, 'last_msg_preview': ''};
    api.handlers['GET /pet-messages'] = () async => {
          'items': [
            _msgJson(1, 'user', '你好'),
            _msgJson(2, 'assistant', '嗨，我在'),
          ],
          'next_cursor': 2,
          'has_more': false,
        };
  });

  tearDown(() async {
    await db.close();
  });

  test('open：本地空 → 拉取服务端历史并渲染', () async {
    await controller.open();

    expect(controller.state.loading, isFalse);
    expect(controller.state.error, isNull);
    expect(controller.state.messages.length, 2);
    expect(controller.state.messages.first.content, '你好');
    expect(controller.state.messages.last.role, PetMessageRole.assistant);
    expect(controller.conversationId, 1);
  });

  test('open：接口失败 → error 为服务端文案', () async {
    api.handlers['POST /pet-conversations'] =
        () async => throw const ApiException(2303, '宠物不存在或无权限');

    await controller.open();

    expect(controller.state.error, '宠物不存在或无权限');
    expect(controller.state.messages, isEmpty);
  });

  test('send：成功后本地出现用户消息与宠物回复', () async {
    await controller.open();
    api.handlers['POST /pet-messages'] =
        () async => _sendResult(clientMsgId: api.lastBody!['client_msg_id'] as String, userContent: '今天很累');

    final ok = await controller.send('今天很累');

    expect(ok, isTrue);
    expect(controller.state.sending, isFalse);
    expect(controller.state.messages.length, 4);
    final last = controller.state.messages.last;
    expect(last.role, PetMessageRole.assistant);
    expect(last.content, '我在这儿呢');
    expect(controller.state.messages.any((m) => m.sendState == PetSendState.sending), isFalse);
  });

  test('send：失败 → 消息标记 failed 且保留重试入口', () async {
    await controller.open();
    api.handlers['POST /pet-messages'] =
        () async => throw const ApiException(2302, '消息内容不合法');

    final ok = await controller.send('   ');

    expect(ok, isFalse); // 空内容直接拦截
    expect(controller.state.messages.length, 2);

    api.handlers['POST /pet-messages'] =
        () async => throw const ApiException(5001, '服务端开小差了');
    final ok2 = await controller.send('你好');

    expect(ok2, isFalse);
    expect(controller.state.error, '服务端开小差了');
    final failed = controller.state.messages.where((m) => m.isFailed).toList();
    expect(failed.length, 1);
    expect(failed.first.content, '你好');
  });

  test('retry：复用同一 client_msg_id 且不新增消息', () async {
    await controller.open();
    api.handlers['POST /pet-messages'] =
        () async => throw const ApiException(5001, '服务端开小差了');
    await controller.send('再试一次');
    final failedClientId = controller.state.messages
        .firstWhere((m) => m.isFailed)
        .clientMsgId!;

    api.handlers['POST /pet-messages'] = () async =>
        _sendResult(clientMsgId: failedClientId, userContent: '再试一次');
    final ok = await controller.retry(failedClientId);

    expect(ok, isTrue);
    expect(api.lastBody!['client_msg_id'], failedClientId);
    expect(controller.state.messages.where((m) => m.isFailed), isEmpty);
    expect(
      controller.state.messages.where((m) => m.content == '再试一次').length,
      1,
      reason: '重试不应产生重复消息',
    );
  });

  test('send：非 ApiException → 通用失败文案', () async {
    await controller.open();
    api.handlers['POST /pet-messages'] = () async => throw Exception('boom');

    final ok = await controller.send('你好');

    expect(ok, isFalse);
    expect(controller.state.error, Zh.chatSendFailed);
  });
}
