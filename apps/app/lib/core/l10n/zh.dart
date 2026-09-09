/// 一期中文文案常量（guide §6.2：所有用户可见文案收敛于此，禁止硬编码在 widget 中）。
class Zh {
  Zh._();

  // 通用
  static const String appName = '语燕';
  static const String confirm = '确认';
  static const String cancel = '取消';
  static const String retry = '重试';

  // 登录
  static const String loginTitle = '登录';
  static const String loginPhoneHint = '手机号';
  static const String loginCodeHint = '验证码';
  static const String loginGetCode = '获取验证码';
  static const String loginSubmit = '登录 / 注册';
  static const String loginSmsSent = '验证码已发送，请查收短信';
  static const String loginCaptchaRefresh = '点击图片刷新验证码';
  static const String loginBadPhone = '请输入正确的 11 位手机号';
  static const String loginEmptyCode = '请输入验证码';

  // 首登引导
  static const String onboardingTitle = '完善资料';
  static const String onboardingPickAvatar = '选择头像';
  static const String onboardingNicknameHint = '昵称（1~20 个字符）';
  static const String onboardingSubmit = '进入语燕';
  static const String onboardingBadNickname = '昵称须为 1~20 个字符';

  // 主页
  static const String homeTitle = '语燕';
  static const String homePlaceholder = '会话列表将在这里出现';
  static const String homeProfile = '我的资料';
  static const String homeSearch = '搜索用户';
  static const String homeContacts = '通讯录';
  static const String homeRequests = '好友申请';

  // 资料页
  static const String profileTitle = '我的资料';
  static const String profileEdit = '编辑资料';
  static const String profileNicknameLabel = '昵称';
  static const String profileAvatarLabel = '头像';
  static const String profileSave = '保存';
  static const String profileLogout = '退出登录';

  // 搜索
  static const String searchTitle = '搜索用户';
  static const String searchHint = '输入对方手机号';
  static const String searchSubmit = '搜索';
  static const String searchEmpty = '未找到该用户';
  static const String searchAddFriend = '加好友';
  static const String searchRequestSent = '已申请';

  // 通讯录
  static const String contactsTitle = '通讯录';
  static const String contactsEmpty = '还没有好友，去搜索添加吧';
  static const String contactsGoAdd = '去添加好友';

  // 好友申请
  static const String requestsTitle = '好友申请';
  static const String requestsEmpty = '暂无新的好友申请';
  static const String requestsAccept = '同意';
  static const String requestsReject = '拒绝';

  // 错误通用
  static const String errorOccurred = '操作失败，请重试';
}
