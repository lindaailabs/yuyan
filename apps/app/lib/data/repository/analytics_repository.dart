import '../remote/api_client.dart';

/// 数据看板：客户端批量上报埋点（POST /events，服务端脱敏）。
class AnalyticsRepository {
  const AnalyticsRepository(this._api);

  final ApiClient _api;

  /// 上报一批埋点，返回服务端接受的条数。
  Future<int> report(List<Map<String, dynamic>> events) async {
    final data = await _api.post(
      '/events',
      body: {'events': events},
    );
    return (data['accepted'] as int?) ?? 0;
  }
}
