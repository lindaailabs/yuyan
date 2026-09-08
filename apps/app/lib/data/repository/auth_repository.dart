import '../model/user_profile.dart';
import '../remote/api_client.dart';

/// 账号域仓库（guide §3.4：widget 只依赖 repository，不触碰 remote 细节）。
class AuthRepository {
  const AuthRepository(this._api);

  final ApiClient _api;

  /// 下发验证码：一期返回图形验证码 base64；生产短信模式 captchaImage 为 null。
  Future<SmsCodeResponse> sendSmsCode(String phone) async {
    final data = await _api.post('/auth/sms-code', body: {'phone': phone});
    return SmsCodeResponse.fromJson(data);
  }

  /// 验证码登录（不存在则自动注册），返回双 token。
  Future<TokenPair> login(String phone, String code) async {
    final data = await _api.post('/auth/login', body: {'phone': phone, 'code': code});
    return TokenPair.fromJson(data);
  }

  /// refresh token 换发新双 token（拦截器内部也用，这里供显式调用）。
  Future<TokenPair> refresh(String refreshToken) async {
    final data = await _api.post('/auth/refresh', body: {'refresh_token': refreshToken});
    return TokenPair.fromJson(data);
  }

  /// 当前用户资料。
  Future<UserProfile> getMe() async {
    final data = await _api.get('/users/me');
    return UserProfile.fromJson(data);
  }

  /// 更新资料（null 字段跳过）。
  Future<UserProfile> updateMe({String? nickname, int? avatarId}) async {
    final body = <String, dynamic>{};
    if (nickname != null) {
      body['nickname'] = nickname;
    }
    if (avatarId != null) {
      body['avatar_id'] = avatarId;
    }
    final data = await _api.put('/users/me', body: body);
    return UserProfile.fromJson(data);
  }

  /// 按手机号精确搜索（脱敏），未注册返回空列表。
  Future<List<UserSearchItem>> search(String phone) async {
    final items = await _api.getList('/users/search', query: {'q': phone});
    return items
        .map((e) => UserSearchItem.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
