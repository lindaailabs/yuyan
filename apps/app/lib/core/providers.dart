import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/remote/api_client.dart';
import '../data/remote/dio_api_client.dart';
import '../data/repository/auth_repository.dart';
import 'auth/auth_controller.dart';
import 'auth/auth_state.dart';
import 'auth/token_storage.dart';

/// API 基地址：默认 Android 模拟器宿主机 loopback；真机/桌面用 --dart-define=API_BASE_URL= 覆盖。
const apiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://10.0.2.2:8080/api/v1',
);

/// Token 安全存储。
final tokenStorageProvider = Provider<TokenStorage>((ref) => SecureTokenStorage());

/// ApiClient：统一包裹解析 + token 注入 + 401 刷新。
/// 显式变量类型：apiClient ↔ authController 存在运行期延迟引用，避免顶层类型推断环。
final Provider<ApiClient> apiClientProvider = Provider<ApiClient>((ref) {
  return DioApiClient(
    baseUrl: apiBaseUrl,
    tokenStorage: ref.watch(tokenStorageProvider),
    // 拦截器回调延迟读取（运行期才触发，无构建期循环依赖）。
    onSessionExpired: () =>
        ref.read(authControllerProvider.notifier).forceLogout(),
  );
});

/// 账号域仓库。
final Provider<AuthRepository> authRepositoryProvider =
    Provider<AuthRepository>((ref) => AuthRepository(ref.watch(apiClientProvider)));

/// 认证状态机（provider 首次被读取时启动恢复流程）。
final StateNotifierProvider<AuthController, AuthState> authControllerProvider =
    StateNotifierProvider<AuthController, AuthState>((ref) {
  final controller = AuthController(
    ref.watch(authRepositoryProvider),
    ref.watch(tokenStorageProvider),
  );
  controller.init();
  return controller;
});
