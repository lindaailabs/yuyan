import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:yuyan_app/core/auth/token_storage.dart';
import 'package:yuyan_app/core/l10n/zh.dart';
import 'package:yuyan_app/core/providers.dart';
import 'package:yuyan_app/main.dart';

void main() {
  testWidgets('YuyanApp：未登录时进入登录页', (WidgetTester tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          apiBaseUrlProvider.overrideWithValue('http://localhost:8080/api/v1'),
          tokenStorageProvider.overrideWithValue(MemoryTokenStorage()),
        ],
        child: const YuyanApp(),
      ),
    );

    await tester.pumpAndSettle();

    expect(find.text(Zh.loginPhoneHint), findsOneWidget);
    expect(find.text(Zh.loginPasswordHint), findsOneWidget);
    expect(find.text(Zh.loginToRegister), findsOneWidget);
  });
}
