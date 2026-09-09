import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/growth/growth_controller.dart';
import 'package:yuyan_app/data/remote/api_exception.dart';
import 'package:yuyan_app/data/repository/pet_growth_repository.dart';

import '../../support/fake_api_client.dart';

Map<String, dynamic> _eventJson(int id, String type, String reason) => {
      'id': id,
      'pet_id': 7,
      'event_type': type,
      'reason': reason,
      'delta': '{"intimacy":2}',
      'created_at': 1788900000 + id,
    };

void main() {
  late FakeApiClient api;
  late PetGrowthRepository repo;
  late GrowthController controller;

  setUp(() {
    api = FakeApiClient();
    repo = PetGrowthRepository(api);
    controller = GrowthController(repo, petId: 7);
  });

  test('load：成功填充时间线', () async {
    api.handlers['GET /pets/7/growth-events'] = () async => {
          'items': [
            _eventJson(2, 'level_up', '亲密度达到 20，从 1 级升到 2 级'),
            _eventJson(1, 'message', '完成一次对话，亲密度 +2'),
          ]
        };

    await controller.load();

    expect(controller.state.loading, isFalse);
    expect(controller.state.error, isNull);
    expect(controller.state.events.length, 2);
    expect(controller.state.events.first.eventType, 'level_up');
    expect(controller.state.events.first.reason, contains('2 级'));
  });

  test('load：ApiException 透传服务端文案', () async {
    api.handlers['GET /pets/7/growth-events'] =
        () async => throw const ApiException(2303, '宠物不存在或无权限');

    await controller.load();

    expect(controller.state.error, '宠物不存在或无权限');
    expect(controller.state.events, isEmpty);
  });

  test('load：非 ApiException 使用通用文案', () async {
    api.handlers['GET /pets/7/growth-events'] =
        () async => throw Exception('boom');

    await controller.load();

    expect(controller.state.error, isNotNull);
  });
}
