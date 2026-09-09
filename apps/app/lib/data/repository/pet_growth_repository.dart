import '../model/pet_growth.dart';
import '../remote/api_client.dart';

/// 成长事件仓库（guide §3.3：UI 不直接调 Dio）。
class PetGrowthRepository {
  const PetGrowthRepository(this._api);

  final ApiClient _api;

  /// 成长时间线（服务端按 id 倒序返回）。
  Future<List<GrowthEvent>> listEvents(int petId, {int? limit}) async {
    final query = <String, dynamic>{};
    if (limit != null) {
      query['limit'] = '$limit';
    }
    final items = await _api.getList(
      '/pets/$petId/growth-events',
      query: query.isEmpty ? null : query,
    );
    return items
        .map((e) => GrowthEvent.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
