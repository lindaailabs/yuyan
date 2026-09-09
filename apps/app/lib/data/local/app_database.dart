import 'package:drift/drift.dart';
import 'package:drift/native.dart';

part 'app_database.g.dart';

/// 本地消息发送状态（客户端私有语义，与服务端 message.status 区分）。
class LocalSendState {
  const LocalSendState._();

  static const int sending = 0;
  static const int sent = 1;
  static const int failed = 2;
}

/// 本地消息缓存：pet_messages 的客户端镜像，用于进入聊天页即时渲染与失败重试。
///
/// 说明：连接当前使用内存库（见下方 _openConnection）。跨进程恢复由服务端历史拉取兜底；
/// 接入 path_provider + sqlite3_flutter_libs 后改为应用文档目录即可获得持久化缓存
/// （属新依赖，需单独评审，不在本变更范围）。
@DataClassName('LocalMessageRow')
class LocalMessages extends Table {
  IntColumn get localId => integer().autoIncrement()();
  IntColumn get convId => integer().named('conv_id')();
  IntColumn get serverId => integer().nullable().named('server_id')();
  TextColumn get role => text().named('role')();
  TextColumn get content => text().named('content')();
  TextColumn get clientMsgId => text().nullable().named('client_msg_id')();
  IntColumn get sendState =>
      integer().named('send_state').withDefault(const Constant(0))();
  IntColumn get errorCode =>
      integer().named('error_code').withDefault(const Constant(0))();
  IntColumn get createdAt => integer().named('created_at')();
}

@DriftAccessor(tables: [LocalMessages])
class MessageDao extends DatabaseAccessor<AppDatabase> with _$MessageDaoMixin {
  MessageDao(super.attachedDatabase);

  /// 会话内的本地消息（按服务端 id 升序，未落库的发送中消息排最后）。
  Future<List<LocalMessageRow>> messagesOf(int convId) {
    return (select(localMessages)
          ..where((t) => t.convId.equals(convId))
          ..orderBy([
            (t) => OrderingTerm.asc(t.serverId),
            (t) => OrderingTerm.asc(t.localId),
          ]))
        .get();
  }

  Future<void> insertMessage(LocalMessagesCompanion row) =>
      into(localMessages).insert(row);

  /// 按 client_msg_id 更新（发送成功后回填服务端 id 与状态）。
  Future<int> updateByClientMsgId(
    String clientMsgId,
    LocalMessagesCompanion row,
  ) =>
      (update(localMessages)
            ..where((t) => t.clientMsgId.equals(clientMsgId)))
          .write(row);

  Future<int> deleteByClientMsgId(String clientMsgId) =>
      (delete(localMessages)..where((t) => t.clientMsgId.equals(clientMsgId)))
          .go();

  /// 会话内已落库的最大服务端 id（增量拉取游标）。
  Future<int> maxServerId(int convId) async {
    final maxExpr = localMessages.serverId.max();
    final row = await (selectOnly(localMessages)
          ..addColumns([maxExpr])
          ..where(localMessages.convId.equals(convId)))
        .getSingleOrNull();
    return row?.read(maxExpr) ?? 0;
  }
}

@DriftDatabase(tables: [LocalMessages], daos: [MessageDao])
class AppDatabase extends _$AppDatabase {
  AppDatabase() : super(_openConnection());

  AppDatabase.test(super.executor);

  @override
  int get schemaVersion => 2;

  @override
  MigrationStrategy get migration => MigrationStrategy(
        onCreate: (Migrator m) async {
          await m.createAll();
        },
        onUpgrade: (Migrator m, int from, int to) async {
          if (from < 2) {
            await m.createTable(localMessages);
          }
        },
      );
}

LazyDatabase _openConnection() {
  return LazyDatabase(() async {
    // 数据库文件路径：接入 path_provider 后改为应用文档目录（见类注释）。
    return NativeDatabase.memory();
  });
}
