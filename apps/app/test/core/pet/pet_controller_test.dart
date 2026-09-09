import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/l10n/zh.dart';
import 'package:yuyan_app/core/pet/pet_controller.dart';
import 'package:yuyan_app/data/remote/api_exception.dart';
import 'package:yuyan_app/data/repository/pet_repository.dart';

import '../../support/fake_api_client.dart';

Map<String, dynamic> _petJson({int id = 1, String name = '小燕'}) => {
  'id': id,
  'user_id': 10,
  'name': name,
  'species': 'swallow',
  'avatar_id': 2,
  'persona': null,
  'level': 1,
  'intimacy': 0,
  'mood': 'curious',
  'created_at': 1788858000,
  'updated_at': 1788858000,
};

void main() {
  late FakeApiClient api;
  late PetController controller;

  setUp(() {
    api = FakeApiClient();
    controller = PetController(PetRepository(api));
  });

  test('load 成功：选择第一只宠物为当前宠物', () async {
    api.handlers['GET /pets'] = () async => {
      'items': [_petJson(id: 1, name: '小燕'), _petJson(id: 2, name: '阿语')],
    };

    await controller.load();

    expect(controller.state.loading, isFalse);
    expect(controller.state.pets.length, 2);
    expect(controller.state.current?.name, '小燕');
    expect(controller.state.error, isNull);
  });

  test('load 失败：显示服务端错误', () async {
    api.handlers['GET /pets'] = () async =>
        throw const ApiException(1002, '凭证无效');

    await controller.load();

    expect(controller.state.loading, isFalse);
    expect(controller.state.error, '凭证无效');
  });

  test('create 成功：新宠物成为当前宠物', () async {
    api.handlers['POST /pets'] = () async => _petJson(id: 3, name: '阿语');

    final ok = await controller.create(name: '阿语', avatarId: 3);

    expect(ok, isTrue);
    expect(controller.state.creating, isFalse);
    expect(controller.state.current?.id, 3);
    expect(controller.state.pets.first.name, '阿语');
  });

  test('create 非 ApiException：显示通用错误', () async {
    api.handlers['POST /pets'] = () async => throw Exception('boom');

    final ok = await controller.create(name: '阿语');

    expect(ok, isFalse);
    expect(controller.state.error, Zh.errorOccurred);
    expect(controller.state.creating, isFalse);
  });
}
