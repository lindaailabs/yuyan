import 'package:drift/drift.dart';
import 'package:drift/native.dart';

// 原生平台：内存 SQLite（接入 path_provider 后可改为文档目录持久化）。
QueryExecutor createConnection() => NativeDatabase.memory();
