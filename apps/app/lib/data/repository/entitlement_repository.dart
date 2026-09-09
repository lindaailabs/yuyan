import '../model/entitlement.dart';
import '../remote/api_client.dart';

/// 权益与额度数据访问：GET /entitlements/me、沙盒开通。
class EntitlementRepository {
  const EntitlementRepository(this._api);

  final ApiClient _api;

  Future<EntitlementView> getMe() async {
    final data = await _api.get('/entitlements/me');
    return EntitlementView.fromJson(data);
  }

  /// 沙盒开通/变更套餐（开发环境可用；生产由服务端返回 2504）。
  Future<EntitlementView> sandboxPurchase(String plan) async {
    final data = await _api.post(
      '/entitlements/sandbox-purchase',
      body: {'plan': plan},
    );
    return EntitlementView.fromJson(data);
  }
}
