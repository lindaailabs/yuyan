// 数据库连接按平台条件选择实现：
// - 原生（Android/iOS/Windows/Linux/macOS）：drift/native（dart:ffi + SQLite）
// - web：drift/web（SQLite WASM，避免 dart:ffi 不可用）
export 'connection_native.dart' if (dart.library.html) 'connection_web.dart';
