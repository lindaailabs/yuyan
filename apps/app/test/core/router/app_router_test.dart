import 'package:flutter_test/flutter_test.dart';
import 'package:yuyan_app/core/auth/auth_state.dart';
import 'package:yuyan_app/core/router/app_router.dart';
import 'package:yuyan_app/data/model/user_profile.dart';

void main() {
  final ready = AuthReady(
    profile: UserProfile(
      id: 1,
      phone: '13811112222',
      nickname: '语燕',
      avatarId: 1,
      createdAt: 1788858000,
    ),
  );

  group('redirectOf（三态守卫纯函数）', () {
    test('LoggedOut：受保护页一律定向 /login，/login 本身放行', () {
      const auth = AuthLoggedOut();
      expect(redirectOf(auth, '/login'), isNull);
      expect(redirectOf(auth, '/home'), '/login');
      expect(redirectOf(auth, '/onboarding'), '/login');
      expect(redirectOf(auth, '/profile'), '/login');
      expect(redirectOf(auth, '/search'), '/login');
    });

    test('Onboarding：未完成引导只能停留在 /onboarding', () {
      const auth = AuthOnboarding();
      expect(redirectOf(auth, '/onboarding'), isNull);
      expect(redirectOf(auth, '/home'), '/onboarding');
      expect(redirectOf(auth, '/login'), '/onboarding');
      expect(redirectOf(auth, '/profile'), '/onboarding');
      expect(redirectOf(auth, '/search'), '/onboarding');
    });

    test('Ready：/login、/onboarding 不可回访，其余放行', () {
      expect(redirectOf(ready, '/login'), '/home');
      expect(redirectOf(ready, '/onboarding'), '/home');
      expect(redirectOf(ready, '/home'), isNull);
      expect(redirectOf(ready, '/profile'), isNull);
      expect(redirectOf(ready, '/search'), isNull);
    });
  });
}
