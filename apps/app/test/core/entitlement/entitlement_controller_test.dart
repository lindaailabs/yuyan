import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/entitlement/entitlement_controller.dart';
import 'package:yuyan_app/data/remote/api_exception.dart';
import 'package:yuyan_app/data/repository/analytics_repository.dart';
import 'package:yuyan_app/data/repository/entitlement_repository.dart';

import '../../support/fake_api_client.dart';

Map<String, dynamic> _entJson(String plan) => {
      'user_id': 10,
      'plan': plan,
      'status': 1,
      'quota': {
        'daily_messages': plan == 'pro' ? 500 : 50,
        'daily_used': 0,
        'daily_remain': plan == 'pro' ? 500 : 50,
        'memory_limit': plan == 'pro' ? 1000 : 200,
        'advanced_model': plan == 'pro',
      },
    };

void main() {
  late FakeApiClient api;
  late EntitlementController controller;

  setUp(() {
    api = FakeApiClient();
    controller = EntitlementController(
      EntitlementRepository(api),
      AnalyticsRepository(api),
    );
  });

  test('load：成功填充权益并上报曝光埋点', () async {
    api.handlers['GET /entitlements/me'] = () async => _entJson('free');
    api.handlers['POST /events'] = () async => {'accepted': 1};

    await controller.load();

    expect(controller.state.loading, isFalse);
    expect(controller.state.error, isNull);
    expect(controller.state.entitlement?.plan, 'free');
    expect(controller.state.entitlement?.quota.dailyRemain, 50);
    // 订阅页曝光埋点已上报。
    expect(api.lastBody?['events'] is List, isTrue);
  });

  test('load：ApiException 透传服务端文案', () async {
    api.handlers['GET /entitlements/me'] =
        () async => throw const ApiException(1002, '未登录');

    await controller.load();

    expect(controller.state.error, '未登录');
    expect(controller.state.entitlement, isNull);
  });

  test('purchase：沙盒开通后切换套餐', () async {
    Map<String, dynamic>? sandboxBody;
    api.handlers['POST /entitlements/sandbox-purchase'] = () async {
      sandboxBody = api.lastBody;
      return _entJson('pro');
    };
    api.handlers['POST /events'] = () async => {'accepted': 1};

    final ok = await controller.purchase('pro');

    expect(ok, isTrue);
    expect(controller.state.entitlement?.isPro, isTrue);
    expect(sandboxBody?['plan'], 'pro');
  });
}
