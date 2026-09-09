import 'package:flutter/foundation.dart';
import 'package:go_router/go_router.dart';

import '../auth/auth_state.dart';
import '../../features/auth/login_page.dart';
import '../../features/auth/onboarding_page.dart';
import '../../features/chat/chat_page.dart';
import '../../features/contacts/contacts_page.dart';
import '../../features/contacts/requests_page.dart';
import '../../features/contacts/search_page.dart';
import '../../features/home/home_page.dart';
import '../../features/pet/create_pet_page.dart';
import '../../features/profile/profile_page.dart';

/// 应用路由（go_router）：三态守卫 redirect。
GoRouter buildRouter(ValueListenable<AuthState> authState) {
  return GoRouter(
    initialLocation: '/login',
    refreshListenable: authState,
    redirect: (context, state) =>
        redirectOf(authState.value, state.matchedLocation),
    routes: [
      GoRoute(path: '/login', builder: (context, state) => const LoginPage()),
      GoRoute(
        path: '/onboarding',
        builder: (context, state) => const OnboardingPage(),
      ),
      GoRoute(path: '/home', builder: (context, state) => const HomePage()),
      GoRoute(
        path: '/pet/create',
        builder: (context, state) => const CreatePetPage(),
      ),
      GoRoute(
        path: '/chat/:petId',
        redirect: (context, state) =>
            int.tryParse(state.pathParameters['petId'] ?? '') == null
                ? '/home'
                : null,
        builder: (context, state) => ChatPage(
          petId: int.parse(state.pathParameters['petId']!),
        ),
      ),
      GoRoute(
        path: '/profile',
        builder: (context, state) => const ProfilePage(),
      ),
      GoRoute(path: '/search', builder: (context, state) => const SearchPage()),
      GoRoute(
        path: '/contacts',
        builder: (context, state) => const ContactsPage(),
      ),
      GoRoute(
        path: '/requests',
        builder: (context, state) => const RequestsPage(),
      ),
    ],
  );
}

/// 路由守卫纯函数：
/// - LoggedOut → /login（且禁止进入受保护页）
/// - Onboarding → /onboarding（未完成引导不得进入主页）
/// - Ready → /home（/login、/onboarding 不可回访）
/// 返回 null 表示放行当前路径。
String? redirectOf(AuthState auth, String location) {
  final inLogin = location == '/login';
  final inOnboarding = location == '/onboarding';

  if (auth is AuthLoggedOut) {
    return inLogin ? null : '/login';
  }
  if (auth is AuthOnboarding) {
    return inOnboarding ? null : '/onboarding';
  }
  // AuthReady：已就绪不得回访登录/引导页。
  return (inLogin || inOnboarding) ? '/home' : null;
}
