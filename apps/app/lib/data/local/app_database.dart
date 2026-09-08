import 'package:drift/drift.dart';
import 'package:drift/native.dart';

part 'app_database.g.dart';

/// App 本地数据库（drift/SQLite，guide §2）。
/// W1 仅建立版本 1 空迁移；W2 起按功能加表（消息、会话等）。
@DriftDatabase(tables: [])
class AppDatabase extends _$AppDatabase {
  AppDatabase() : super(_openConnection());

  AppDatabase.test(super.executor);

  @override
  int get schemaVersion => 1;

  @override
  MigrationStrategy get migration => MigrationStrategy(
        onCreate: (Migrator m) async {
          // W1 无业务表；后续版本的表在此按版本追加。
        },
      );
}

LazyDatabase _openConnection() {
  return LazyDatabase(() async {
    // 数据库文件路径：W2 接入 path_provider 后改为应用文档目录。
    return NativeDatabase.memory();
  });
}
