import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/memory/memory_controller.dart';
import 'package:yuyan_app/data/remote/api_exception.dart';
import 'package:yuyan_app/data/repository/pet_memory_repository.dart';

import '../../support/fake_api_client.dart';

Map<String, dynamic> _memoryJson(int id, String content) => {
      'id': id,
      'pet_id': 7,
      'type': 'preference',
      'content': content,
      'confidence': 0.8,
      'created_at': 1788900000 + id,
    };

void main() {
  late FakeApiClient api;
  late PetMemoryRepository repo;
  late MemoryController controller;

  setUp(() {
    api = FakeApiClient();
    repo = PetMemoryRepository(api);
    controller = MemoryController(repo, petId: 7);
  });

  test('load：成功填充列表', () async {
    api.handlers['GET /pets/7/memories'] = () async => {
          'items': [_memoryJson(1, '用户喜欢蓝色'), _memoryJson(2, '用户住在杭州')]
        };

    await controller.load();

    expect(controller.state.loading, isFalse);
    expect(controller.state.error, isNull);
    expect(controller.state.memories.length, 2);
    expect(controller.state.memories.first.content, '用户喜欢蓝色');
    expect(controller.state.memories.first.type, 'preference');
  });

  test('load：ApiException 透传服务端文案', () async {
    api.handlers['GET /pets/7/memories'] =
        () async => throw const ApiException(2303, '宠物不存在或无权限');

    await controller.load();

    expect(controller.state.error, '宠物不存在或无权限');
    expect(controller.state.memories, isEmpty);
  });

  test('load：非 ApiException 用通用文案', () async {
    api.handlers['GET /pets/7/memories'] = () async => throw Exception('boom');

    await controller.load();

    expect(controller.state.error, isNotNull);
  });

  test('delete：成功后从列表局部移除', () async {
    api.handlers['GET /pets/7/memories'] = () async => {
          'items': [_memoryJson(1, '用户喜欢蓝色'), _memoryJson(2, '用户住在杭州')]
        };
    await controller.load();

    var deleted = false;
    api.handlers['DELETE /pet-memories/1'] = () async {
      deleted = true;
      return {};
    };

    final ok = await controller.delete(1);

    expect(ok, isTrue);
    expect(deleted, isTrue);
    expect(controller.state.memories.length, 1);
    expect(controller.state.memories.single.id, 2);
    expect(controller.state.deletingIds, isEmpty);
    expect(controller.state.error, isNull);
  });

  test('delete：失败保留列表并给出文案', () async {
    api.handlers['GET /pets/7/memories'] = () async => {
          'items': [_memoryJson(1, '用户喜欢蓝色')]
        };
    await controller.load();
    api.handlers['DELETE /pet-memories/1'] =
        () async => throw const ApiException(2401, '记忆不存在或无权限');

    final ok = await controller.delete(1);

    expect(ok, isFalse);
    expect(controller.state.memories.length, 1);
    expect(controller.state.error, '记忆不存在或无权限');
    expect(controller.state.deletingIds, isEmpty);
  });
}
