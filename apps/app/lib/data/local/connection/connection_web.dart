// web 调试用 sql.js 后端（drift/web.dart 在 2.34 已标记弃用，但 web 仅为本地调试通道，
// 正式客户端走原生 drift/native，故此处忽略弃用告警）。
// ignore_for_file: deprecated_member_use
import 'package:drift/drift.dart';
import 'package:drift/web.dart';

// web 平台：drift 2.34.x 的 WebDatabase 走 sql.js（见 web/index.html 引入）。
// 数据为内存/IndexedDB，重启即丢，仅用于浏览器联调，不验证持久化。
QueryExecutor createConnection() => WebDatabase('yuyan');
