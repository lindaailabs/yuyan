import 'package:drift/drift.dart';
import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/data/local/app_database.dart';

void main() {
  late AppDatabase db;
  late MessageDao dao;

  setUp(() {
    db = AppDatabase.test(NativeDatabase.memory());
    dao = MessageDao(db);
  });

  tearDown(() async {
    await db.close();
  });

  LocalMessagesCompanion row({
    required int convId,
    int? serverId,
    required String role,
    required String content,
    String? clientMsgId,
    int sendState = LocalSendState.sent,
    int createdAt = 1788900000,
  }) =>
      LocalMessagesCompanion.insert(
        convId: convId,
        serverId: serverId == null ? const Value.absent() : Value(serverId),
        role: role,
        content: content,
        clientMsgId: Value(clientMsgId),
        sendState: Value(sendState),
        createdAt: createdAt,
      );

  test('插入与按会话读取（服务端 id 升序）', () async {
    await dao.insertMessage(row(convId: 1, serverId: 2, role: 'assistant', content: 'b'));
    await dao.insertMessage(row(convId: 1, serverId: 1, role: 'user', content: 'a'));
    await dao.insertMessage(row(convId: 2, serverId: 3, role: 'user', content: 'x'));

    final list = await dao.messagesOf(1);
    expect(list.length, 2);
    expect(list.first.serverId, 1);
    expect(list.last.serverId, 2);
  });

  test('maxServerId 作为增量游标，忽略未落库的发送中消息', () async {
    await dao.insertMessage(row(convId: 1, serverId: 5, role: 'user', content: 'a'));
    await dao.insertMessage(
      row(convId: 1, role: 'user', content: 'sending', clientMsgId: 'cid-1', sendState: LocalSendState.sending),
    );

    expect(await dao.maxServerId(1), 5);
    expect(await dao.maxServerId(9), 0);
  });

  test('按 client_msg_id 更新为已发送（回填服务端 id）', () async {
    await dao.insertMessage(
      row(convId: 1, role: 'user', content: 'hi', clientMsgId: 'cid-2', sendState: LocalSendState.sending),
    );

    final affected = await dao.updateByClientMsgId(
      'cid-2',
      const LocalMessagesCompanion(
        serverId: Value(11),
        sendState: Value(LocalSendState.sent),
      ),
    );

    expect(affected, 1);
    final list = await dao.messagesOf(1);
    expect(list.single.serverId, 11);
    expect(list.single.sendState, LocalSendState.sent);
  });

  test('按 client_msg_id 删除', () async {
    await dao.insertMessage(row(convId: 1, serverId: 1, role: 'user', content: 'a', clientMsgId: 'cid-3'));
    expect(await dao.deleteByClientMsgId('cid-3'), 1);
    expect(await dao.messagesOf(1), isEmpty);
  });
}
