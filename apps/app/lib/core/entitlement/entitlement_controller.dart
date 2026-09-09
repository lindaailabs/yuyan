import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/model/entitlement.dart';
import '../../data/remote/api_exception.dart';
import '../../data/repository/analytics_repository.dart';
import '../../data/repository/entitlement_repository.dart';
import '../l10n/zh.dart';

/// 权益页状态：加载 / 当前权益 / 错误。
class EntitlementState {
  const EntitlementState({
    this.entitlement,
    this.loading = false,
    this.purchasing = false,
    this.error,
  });

  final EntitlementView? entitlement;
  final bool loading;
  final bool purchasing;
  final String? error;

  EntitlementState copyWith({
    EntitlementView? entitlement,
    bool? loading,
    bool? purchasing,
    String? error,
    bool clearError = false,
  }) =>
      EntitlementState(
        entitlement: entitlement ?? this.entitlement,
        loading: loading ?? this.loading,
        purchasing: purchasing ?? this.purchasing,
        error: clearError ? null : error ?? this.error,
      );
}

/// 权益状态机：加载当前权益、订阅页曝光埋点、沙盒开通/变更套餐。
class EntitlementController extends StateNotifier<EntitlementState> {
  EntitlementController(this._entRepo, this._analyticsRepo)
      : super(const EntitlementState());

  final EntitlementRepository _entRepo;
  final AnalyticsRepository _analyticsRepo;

  Future<void> load() async {
    state = state.copyWith(loading: true, clearError: true);
    try {
      final view = await _entRepo.getMe();
      state = state.copyWith(entitlement: view, loading: false);
      // 订阅页曝光埋点（服务端脱敏）。
      await _report(const {'name': 'subscription_view'});
    } on ApiException catch (e) {
      state = state.copyWith(loading: false, error: e.msg);
    } catch (_) {
      state = state.copyWith(loading: false, error: Zh.errorOccurred);
    }
  }

  Future<bool> purchase(String plan) async {
    if (state.purchasing) return false;
    state = state.copyWith(purchasing: true, clearError: true);
    try {
      final view = await _entRepo.sandboxPurchase(plan);
      state = state.copyWith(entitlement: view, purchasing: false);
      await _report({
        'name': 'subscription_purchase',
        'props': {'plan': plan, 'channel': 'sandbox'},
      });
      return true;
    } on ApiException catch (e) {
      state = state.copyWith(purchasing: false, error: e.msg);
      return false;
    } catch (_) {
      state = state.copyWith(purchasing: false, error: Zh.errorOccurred);
      return false;
    }
  }

  Future<void> _report(Map<String, dynamic> event) async {
    try {
      await _analyticsRepo.report([event]);
    } catch (_) {
      // 埋点失败不影响主流程。
    }
  }
}
