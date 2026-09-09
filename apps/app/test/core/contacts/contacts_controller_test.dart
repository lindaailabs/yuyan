import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/contacts/contacts_controller.dart';
import 'package:yuyan_app/core/l10n/zh.dart';
import 'package:yuyan_app/data/remote/api_exception.dart';
import 'package:yuyan_app/data/repository/contacts_repository.dart';

import '../../support/fake_api_client.dart';

Map<String, dynamic> _userJson(int id, String nickname) => {
      'id': id,
      'phone': '138****000$id',
      'nickname': nickname,
      'avatar_id': id,
    };

Map<String, dynamic> _requestJson(int id, int fromId) => {
      'id': id,
      'from_user': _userJson(fromId, '用户$fromId'),
      'created_at': 1788858000 + id,
    };

Map<String, dynamic> _friendJson(int userId) => {
      'user': _userJson(userId, '用户$userId'),
      'created_at': 1788858000 + userId,
    };

void main() {
  late FakeApiClient api;
  late ContactsRepository repo;
  late ContactsController controller;

  setUp(() {
    api = FakeApiClient();
    repo = ContactsRepository(api);
    controller = ContactsController(repo);
  });

  group('loadAll', () {
    test('成功：双列表填充，loading 复位且无错误', () async {
      api.handlers['GET /friends'] = () async =>
          {'items': [_friendJson(2), _friendJson(3)]};
      api.handlers['GET /friends/requests'] = () async =>
          {'items': [_requestJson(10, 4)]};

      await controller.loadAll();

      expect(controller.state.friends.length, 2);
      expect(controller.state.friends[0].user.nickname, '用户2');
      expect(controller.state.requests.length, 1);
      expect(controller.state.requests[0].fromUser.id, 4);
      expect(controller.state.loading, isFalse);
      expect(controller.state.error, isNull);
    });

    test('ApiException：error 为服务端 msg，列表保持为空', () async {
      api.handlers['GET /friends'] =
          () async => throw const ApiException(1002, '凭证无效');
      api.handlers['GET /friends/requests'] = () async => {'items': []};

      await controller.loadAll();

      expect(controller.state.error, '凭证无效');
      expect(controller.state.friends, isEmpty);
      expect(controller.state.loading, isFalse);
    });

    test('非 ApiException：error 为通用文案', () async {
      api.handlers['GET /friends'] = () async => throw Exception('boom');
      api.handlers['GET /friends/requests'] = () async => {'items': []};

      await controller.loadAll();

      expect(controller.state.error, Zh.errorOccurred);
    });
  });

  group('accept', () {
    test('成功：申请移除 + 好友追加到列表末尾', () async {
      api.handlers['GET /friends'] = () async =>
          {'items': [_friendJson(2)]};
      api.handlers['GET /friends/requests'] = () async =>
          {'items': [_requestJson(10, 4)]};
      await controller.loadAll();

      api.handlers['POST /friends/requests/10/accept'] = () async => {
            'id': 10,
            'status': 2,
          };

      final ok = await controller.accept(10);

      expect(ok, isTrue);
      expect(controller.state.requests, isEmpty);
      expect(controller.state.friends.length, 2);
      expect(controller.state.friends.last.user.id, 4);
      expect(controller.state.processingIds, isEmpty);
      expect(controller.state.error, isNull);
    });

    test('ApiException：列表不变，error 为服务端 msg', () async {
      api.handlers['GET /friends'] = () async => {'items': []};
      api.handlers['GET /friends/requests'] = () async =>
          {'items': [_requestJson(10, 4)]};
      await controller.loadAll();

      api.handlers['POST /friends/requests/10/accept'] =
          () async => throw const ApiException(2106, '申请不存在或已处理');

      final ok = await controller.accept(10);

      expect(ok, isFalse);
      expect(controller.state.requests.length, 1);
      expect(controller.state.friends, isEmpty);
      expect(controller.state.error, '申请不存在或已处理');
      expect(controller.state.processingIds, isEmpty);
    });

    test('非 ApiException：error 为通用文案', () async {
      api.handlers['GET /friends'] = () async => {'items': []};
      api.handlers['GET /friends/requests'] = () async =>
          {'items': [_requestJson(10, 4)]};
      await controller.loadAll();

      api.handlers['POST /friends/requests/10/accept'] =
          () async => throw Exception('boom');

      final ok = await controller.accept(10);

      expect(ok, isFalse);
      expect(controller.state.error, Zh.errorOccurred);
    });
  });

  group('reject', () {
    test('成功：仅申请移除，好友列表不变', () async {
      api.handlers['GET /friends'] = () async =>
          {'items': [_friendJson(2)]};
      api.handlers['GET /friends/requests'] = () async =>
          {'items': [_requestJson(10, 4), _requestJson(11, 5)]};
      await controller.loadAll();

      api.handlers['POST /friends/requests/10/reject'] = () async => {
            'id': 10,
            'status': 4,
          };

      final ok = await controller.reject(10);

      expect(ok, isTrue);
      expect(controller.state.requests.length, 1);
      expect(controller.state.requests.first.id, 11);
      expect(controller.state.friends.length, 1);
      expect(controller.state.processingIds, isEmpty);
    });

    test('ApiException：申请保留，error 为服务端 msg', () async {
      api.handlers['GET /friends'] = () async => {'items': []};
      api.handlers['GET /friends/requests'] = () async =>
          {'items': [_requestJson(10, 4)]};
      await controller.loadAll();

      api.handlers['POST /friends/requests/10/reject'] =
          () async => throw const ApiException(2106, '申请不存在或已处理');

      final ok = await controller.reject(10);

      expect(ok, isFalse);
      expect(controller.state.requests.length, 1);
      expect(controller.state.error, '申请不存在或已处理');
    });
  });

  group('sendRequest', () {
    test('成功：透传 POST /friends/requests，无状态变更', () async {
      var posted = false;
      api.handlers['POST /friends/requests'] = () async {
        posted = true;
        return {'id': 9, 'user_id': 1, 'friend_id': 4, 'status': 1, 'created_at': 0};
      };

      await controller.sendRequest(4);

      expect(posted, isTrue);
      expect(controller.state.error, isNull);
    });

    test('业务错误：ApiException 原样上抛（UI snackbar 呈现）', () async {
      api.handlers['POST /friends/requests'] =
          () async => throw const ApiException(2104, '你们已经是好友');

      await expectLater(
        controller.sendRequest(4),
        throwsA(isA<ApiException>()
            .having((e) => e.msg, 'msg', '你们已经是好友')),
      );
    });
  });
}
