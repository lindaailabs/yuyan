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
  static const String loginPasswordHint = '密码';
  static const String loginSubmit = '登录';
  static const String loginBadPhone = '请输入正确的 11 位手机号';
  static const String loginEmptyPassword = '请输入密码';
  static const String loginToRegister = '没有账号？去注册';

  // 注册
  static const String registerTitle = '注册';
  static const String registerPhoneHint = '手机号';
  static const String registerPasswordHint = '密码（6~64 位）';
  static const String registerConfirmHint = '确认密码';
  static const String registerSubmit = '注册并登录';
  static const String registerToLogin = '已有账号？去登录';
  static const String registerPasswordMismatch = '两次输入的密码不一致';
  static const String registerBadPhone = '请输入正确的 11 位手机号';
  static const String registerBadPassword = '密码需为 6~64 位字符';

  // 首登引导
  static const String onboardingTitle = '完善资料';
  static const String onboardingPickAvatar = '选择头像';
  static const String onboardingNicknameHint = '昵称（1~20 个字符）';
  static const String onboardingSubmit = '进入语燕';
  static const String onboardingBadNickname = '昵称须为 1~20 个字符';

  // 宠物主页
  static const String homeTitle = '语燕';
  static const String homeProfile = '我的资料';
  static const String homePetCreate = '领养宠物';
  static const String homePetEmptyTitle = '还没有宠物';
  static const String homePetEmptyBody = '先领养一只会记得你、会成长的 AI 宠物。';
  static const String homePetMood = '心情';
  static const String homePetLevel = '等级';
  static const String homePetIntimacy = '亲密度';
  static const String homePetChat = '对话';
  static const String homePetMemory = '记忆';
  static const String homePetGrowth = '成长';
  static const String homePetChatComing = '文本对话将在 W2 接入';
  static const String homePetMemoryComing = '记忆系统将在 W3 接入';
  static const String homePetGrowthComing = '成长事件将在 W4 接入';

  // 宠物切换（多只宠物）
  static const String petSwitch = '切换宠物';
  static const String petSwitchTitle = '选择要陪伴的宠物';

  // 宠物创建
  static const String petCreateTitle = '领养宠物';
  static const String petNameHint = '给它起个名字';
  static const String petPickAvatar = '选择外观';
  static const String petCreateSubmit = '开始陪伴';
  static const String petBadName = '宠物名字须为 1~20 个字符';

  // 资料页
  static const String profileTitle = '我的资料';
  static const String profileEdit = '编辑资料';
  static const String profileNicknameLabel = '昵称';
  static const String profileAvatarLabel = '头像';
  static const String profileSave = '保存';
  static const String profileLogout = '退出登录';

  // 搜索（早期 IM 遗留入口）
  static const String homeSearch = '搜索用户';
  static const String homeContacts = '通讯录';
  static const String homeRequests = '好友申请';
  static const String searchTitle = '搜索用户';
  static const String searchHint = '输入对方手机号';
  static const String searchSubmit = '搜索';
  static const String searchEmpty = '未找到该用户';
  static const String searchAddFriend = '加好友';
  static const String searchRequestSent = '已申请';

  // 通讯录（早期 IM 遗留入口）
  static const String contactsTitle = '通讯录';
  static const String contactsEmpty = '还没有好友，去搜索添加吧';
  static const String contactsGoAdd = '去添加好友';

  // 好友申请（早期 IM 遗留入口）
  static const String requestsTitle = '好友申请';
  static const String requestsEmpty = '暂无新的好友申请';
  static const String requestsAccept = '同意';
  static const String requestsReject = '拒绝';

  // 聊天
  static const String chatGreeting = '嗨，我在这儿陪你';
  static const String chatEmptyHint = '说点什么开始我们的第一次对话吧';
  static const String chatInputHint = '说点什么…';
  static const String chatChip1 = '今天过得怎么样？';
  static const String chatChip2 = '陪我聊聊天';
  static const String chatChip3 = '讲个小故事';
  static const String chatThinking = '正在思考…';
  static const String chatFailedHint = '发送失败，点击重试';
  static const String chatSendFailed = '消息没发出去，请重试';
  static const String chatLoadMore = '加载更多';
  static const String chatSyncing = '正在同步新消息…';

  // 记忆
  static const String memoryTitle = '它记住的事';
  static const String memoryEmpty = '还没有记住任何事';
  static const String memoryEmptyHint = '和它多聊聊，我会把你的喜好和习惯记下来';
  static const String memoryTypeLabel = '类型';
  static const String memoryConfidence = '置信度';
  static const String memoryDeleteTitle = '删除这条记忆？';
  static const String memoryDeleteBody = '删除后它将不再记得这件事，也不会再用它来陪你聊天。';
  static const String memoryDeleteConfirm = '删除';
  static const String memoryDeleteFailed = '删除失败，请重试';
  static const String chatNewMemory = '记住了';

  // 成长
  static const String growthTitle = '成长记录';
  static const String growthEmpty = '还没有成长记录';
  static const String growthEmptyHint = '多陪它说说话，每一次变化都会记在这里';
  static const String growthTypeMessage = '一次陪伴';
  static const String growthTypeLevelUp = '升级啦';
  static const String growthTypeMood = '心情变化';
  static const String growthTypeStreak = '连续互动';

  // 订阅 / 权益
  static const String subscription = '订阅与权益';
  static const String planFree = '免费版';
  static const String planFreeSub = '每天 50 条 AI 对话，记录 200 条长期记忆';
  static const String planPro = 'Pro 版';
  static const String planProSub = '每天 500 条 AI 对话，记录 1000 条长期记忆';
  static const String quotaDaily = '今日 AI 对话额度';
  static const String quotaDailyRemain = '今日剩余';
  static const String quotaMemory = '长期记忆上限';
  static const String quotaAdvancedModel = '高级模型';
  static const String enabled = '已开启';
  static const String disabled = '未开启';
  static const String subscriptionSandbox = '沙盒开通（开发环境）';
  static const String subscriptionUpgradePro = '升级到 Pro';
  static const String subscriptionRenewPro = '续费 Pro';
  static const String subscriptionBackFree = '恢复免费版';
  static const String subscriptionSandboxHint = '当前为开发环境，可用沙盒直接切换套餐；正式支付请走真实回调。';
  static const String goSubscribe = '去订阅';
  static const String chatQuotaExhausted = '今天的 AI 对话额度用完啦，升级 Pro 可继续畅聊。';

  // 错误通用
  static const String errorOccurred = '操作失败，请重试';
}
