import 'token_storage_base.dart';
import 'token_storage_io.dart'
    if (dart.library.html) 'token_storage_web.dart'
    as platform;

export 'token_storage_base.dart';

/// 平台 token 存储。
///
/// - Android/iOS/桌面：flutter_secure_storage。
/// - Web：优先 flutter_secure_storage；非安全来源下退到 sessionStorage，方便
///   `http://192.168.x.x` 本地联调。
TokenStorage createTokenStorage() => platform.createTokenStorage();
