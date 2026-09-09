import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/data/local/app_database.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test('AppDatabase 以内存执行器打开并可关闭（版本 2 含本地消息表）', () async {
    final db = AppDatabase.test(NativeDatabase.memory());
    try {
      final ver = await db.customSelect('PRAGMA user_version').getSingle();
      expect(ver.data.values.first, 2); // drift 将 schemaVersion 写入 PRAGMA
      expect(db.schemaVersion, 2);
    } finally {
      await db.close();
    }
  });
}
