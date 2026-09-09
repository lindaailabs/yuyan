import 'dart:convert';

import 'package:flutter/services.dart';

/// 应用配置：接口前缀等按环境集中管理，对应后端的 config.test.yaml / config.prod.yaml。
///
/// 取值优先级（高→低）：
/// 1. 编译期注入 `--dart-define=API_BASE_URL=...`（一次性临时覆盖 URL）
/// 2. `--dart-define=APP_ENV=<env>` 指定的 `assets/config.<env>.json`（与后端 APP_ENV 对应）
/// 3. 默认 `assets/config.json`
/// 4. 内置默认（Android 模拟器宿主机 loopback）
///
/// 本地调试：直接 `flutter run -d chrome`（APP_ENV 默认 test，加载 config.test.json，零参数）。
/// 正式打包：只需短参数 `flutter build apk --dart-define=APP_ENV=prod`（加载 config.prod.json）。
class AppConfig {
  static late final String apiBaseUrl;

  /// 当前环境：默认 test（对应后端 config.test.yaml）。
  static const String env =
      String.fromEnvironment('APP_ENV', defaultValue: 'test');

  static Future<void> load() async {
    // 1) 编译期注入的 URL 优先级最高。
    const fromDefine = String.fromEnvironment('API_BASE_URL', defaultValue: '');
    if (fromDefine.isNotEmpty) {
      apiBaseUrl = fromDefine;
      return;
    }
    // 2) 按环境读 assets/config.<env>.json。
    final fromEnvFile = await _readFile('assets/config.$env.json');
    if (fromEnvFile != null && fromEnvFile.isNotEmpty) {
      apiBaseUrl = fromEnvFile;
      return;
    }
    // 3) 默认文件。
    final fromDefaultFile = await _readFile('assets/config.json');
    if (fromDefaultFile != null && fromDefaultFile.isNotEmpty) {
      apiBaseUrl = fromDefaultFile;
      return;
    }
    // 4) 内置默认。
    apiBaseUrl = 'http://10.0.2.2:8080/api/v1';
  }

  static Future<String?> _readFile(String path) async {
    try {
      final raw = await rootBundle.loadString(path);
      final map = jsonDecode(raw) as Map<String, dynamic>;
      final value = map['apiBaseUrl'];
      return value is String && value.isNotEmpty ? value : null;
    } catch (_) {
      // 文件缺失/解析失败：返回 null 继续回退。
      return null;
    }
  }
}
