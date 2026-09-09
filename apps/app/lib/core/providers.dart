import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'app_config.dart';
import '../data/local/app_database.dart';
import '../data/remote/api_client.dart';
import '../data/remote/dio_api_client.dart';
import '../data/repository/analytics_repository.dart';
import '../data/repository/auth_repository.dart';
import '../data/repository/contacts_repository.dart';
import '../data/repository/entitlement_repository.dart';
import '../data/repository/pet_chat_repository.dart';
import '../data/repository/pet_growth_repository.dart';
import '../data/repository/pet_memory_repository.dart';
import '../data/repository/pet_repository.dart';
import 'auth/auth_controller.dart';
import 'auth/auth_state.dart';
import 'auth/token_storage.dart';
import 'chat/chat_controller.dart';
import 'contacts/contacts_controller.dart';
import 'entitlement/entitlement_controller.dart';
import 'growth/growth_controller.dart';
import 'memory/memory_controller.dart';
import 'pet/pet_controller.dart';

/// API 基地址（统一前缀）：加载顺序 --dart-define=API_BASE_URL > assets/config.json > 内置默认。
/// 所有接口共用此前缀，改 assets/config.json 即可切换调试/打包地址，不必每次 build 都带 --dart-define。
/// 见 core/app_config.dart。
final apiBaseUrlProvider = Provider<String>((ref) => AppConfig.apiBaseUrl);

/// Token 安全存储。
final tokenStorageProvider = Provider<TokenStorage>(
  (ref) => SecureTokenStorage(),
);

/// ApiClient：统一包裹解析 + token 注入 + 401 刷新。
/// 显式变量类型：apiClient ↔ authController 存在运行期延迟引用，避免顶层类型推断环。
final Provider<ApiClient> apiClientProvider = Provider<ApiClient>((ref) {
  return DioApiClient(
    baseUrl: ref.watch(apiBaseUrlProvider),
    tokenStorage: ref.watch(tokenStorageProvider),
    // 拦截器回调延迟读取（运行期才触发，无构建期循环依赖）。
    onSessionExpired: () =>
        ref.read(authControllerProvider.notifier).forceLogout(),
  );
});

/// 账号域仓库。
final Provider<AuthRepository> authRepositoryProvider =
    Provider<AuthRepository>(
      (ref) => AuthRepository(ref.watch(apiClientProvider)),
    );

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

/// 本地库（drift）：进程内缓存消息与宠物状态。
final Provider<AppDatabase> appDatabaseProvider = Provider<AppDatabase>((ref) {
  final db = AppDatabase();
  ref.onDispose(db.close);
  return db;
});

/// 本地消息 DAO。
final Provider<MessageDao> messageDaoProvider = Provider<MessageDao>(
  (ref) => MessageDao(ref.watch(appDatabaseProvider)),
);

/// 宠物对话仓库（远端 + 本地缓存合并）。
final Provider<PetChatRepository> petChatRepositoryProvider =
    Provider<PetChatRepository>(
      (ref) => PetChatRepository(
        ref.watch(apiClientProvider),
        ref.watch(messageDaoProvider),
      ),
    );

/// 聊天页状态机（按宠物 id 分实例，进入时加载本地并增量同步）。
final StateNotifierProviderFamily<ChatController, ChatState, int>
chatControllerProvider =
    StateNotifierProvider.family<ChatController, ChatState, int>((ref, petId) {
      final controller = ChatController(
        ref.watch(petChatRepositoryProvider),
        petId: petId,
      );
      controller.open();
      return controller;
    });

/// 记忆仓库（查看与删除长期记忆）。
final Provider<PetMemoryRepository> petMemoryRepositoryProvider =
    Provider<PetMemoryRepository>(
      (ref) => PetMemoryRepository(ref.watch(apiClientProvider)),
    );

/// 记忆页状态机（按宠物 id 分实例）。
final StateNotifierProviderFamily<MemoryController, MemoryState, int>
memoryControllerProvider =
    StateNotifierProvider.family<MemoryController, MemoryState, int>(
        (ref, petId) {
      return MemoryController(
        ref.watch(petMemoryRepositoryProvider),
        petId: petId,
      );
    });

/// 成长事件仓库。
final Provider<PetGrowthRepository> petGrowthRepositoryProvider =
    Provider<PetGrowthRepository>(
      (ref) => PetGrowthRepository(ref.watch(apiClientProvider)),
    );

/// 成长时间线状态机（按宠物 id 分实例）。
final StateNotifierProviderFamily<GrowthController, GrowthState, int>
growthControllerProvider =
    StateNotifierProvider.family<GrowthController, GrowthState, int>(
        (ref, petId) {
      return GrowthController(
        ref.watch(petGrowthRepositoryProvider),
        petId: petId,
      );
    });

/// 宠物域仓库。
final Provider<PetRepository> petRepositoryProvider = Provider<PetRepository>(
  (ref) => PetRepository(ref.watch(apiClientProvider)),
);

/// 宠物主页状态机。
final StateNotifierProvider<PetController, PetHomeState> petControllerProvider =
    StateNotifierProvider<PetController, PetHomeState>((ref) {
      final controller = PetController(ref.watch(petRepositoryProvider));
      controller.load();
      return controller;
    });

/// 好友域仓库（早期 IM 遗留能力，非一期 AI 宠物主线）。
final Provider<ContactsRepository> contactsRepositoryProvider =
    Provider<ContactsRepository>(
      (ref) => ContactsRepository(ref.watch(apiClientProvider)),
    );

/// 好友域状态机（仅旧页面使用，主页不再预载）。
final StateNotifierProvider<ContactsController, ContactsState>
contactsControllerProvider =
    StateNotifierProvider<ContactsController, ContactsState>((ref) {
      final controller = ContactsController(
        ref.watch(contactsRepositoryProvider),
      );
      controller.loadAll();
      return controller;
    });

/// 权益与额度仓库。
final Provider<EntitlementRepository> entitlementRepositoryProvider =
    Provider<EntitlementRepository>(
      (ref) => EntitlementRepository(ref.watch(apiClientProvider)),
    );

/// 数据看板仓库（埋点上报）。
final Provider<AnalyticsRepository> analyticsRepositoryProvider =
    Provider<AnalyticsRepository>(
      (ref) => AnalyticsRepository(ref.watch(apiClientProvider)),
    );

/// 权益页状态机（单一实例）。
final StateNotifierProvider<EntitlementController, EntitlementState>
entitlementControllerProvider =
    StateNotifierProvider<EntitlementController, EntitlementState>((ref) {
      return EntitlementController(
        ref.watch(entitlementRepositoryProvider),
        ref.watch(analyticsRepositoryProvider),
      );
    });
