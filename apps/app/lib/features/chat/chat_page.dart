import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/auth/auth_state.dart';
import '../../core/chat/chat_controller.dart';
import '../../core/l10n/zh.dart';
import '../../core/providers.dart';
import '../../data/model/pet_memory.dart';
import '../../data/model/pet_message.dart';
import '../shared/avatar_widget.dart';

/// 设计令牌（奶油白底 + 淡紫主色 + 暖橙点缀）。
const _bg = Color(0xFFFFF9F5);
const _primary = Color(0xFF7C6BF5);
const _primarySoft = Color(0xFF9C8CFF);
const _accent = Color(0xFFFFB86B);
const _ink = Color(0xFF2B2B34);
const _muted = Color(0xFF7A7A88);
const _danger = Color(0xFFE5484D);

/// 宠物聊天页：本地优先渲染 → 增量同步；发送中/失败可重试。
class ChatPage extends ConsumerStatefulWidget {
  const ChatPage({super.key, required this.petId});

  final int petId;

  @override
  ConsumerState<ChatPage> createState() => _ChatPageState();
}

class _ChatPageState extends ConsumerState<ChatPage>
    with SingleTickerProviderStateMixin {
  final _inputCtrl = TextEditingController();
  final _scrollCtrl = ScrollController();
  late final AnimationController _breathCtrl;
  late final Animation<double> _breath;

  @override
  void initState() {
    super.initState();
    _breathCtrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 2200),
    )..repeat(reverse: true);
    _breath = Tween<double>(
      begin: 0.97,
      end: 1.05,
    ).animate(CurvedAnimation(parent: _breathCtrl, curve: Curves.easeInOut));
    _inputCtrl.addListener(() => setState(() {}));
  }

  @override
  void dispose() {
    _inputCtrl.dispose();
    _scrollCtrl.dispose();
    _breathCtrl.dispose();
    super.dispose();
  }

  Future<void> _send() async {
    final text = _inputCtrl.text;
    if (text.trim().isEmpty) return;
    _inputCtrl.clear();
    final ok = await ref
        .read(chatControllerProvider(widget.petId).notifier)
        .send(text);
    if (!ok || !mounted) return;
    _scrollToBottom();
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!_scrollCtrl.hasClients) return;
      _scrollCtrl.animateTo(
        0,
        duration: const Duration(milliseconds: 260),
        curve: Curves.easeOut,
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(chatControllerProvider(widget.petId));
    final petState = ref.watch(petControllerProvider);
    final pet = petState.current;
    final petName = pet != null && pet.id == widget.petId ? pet.name : '宠物';
    final authState = ref.watch(authControllerProvider);
    final userAvatarId =
        authState is AuthReady ? authState.profile.avatarId : 1;
    final petAvatarId = pet?.avatarId ?? 1;

    return Scaffold(
      backgroundColor: _bg,
      appBar: AppBar(
        backgroundColor: _bg,
        surfaceTintColor: Colors.transparent,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new_rounded),
          onPressed: () => Navigator.of(context).maybePop(),
        ),
        title: Row(
          children: [
            ScaleTransition(
              scale: _breath,
              child: PetAvatarWidget(avatarId: pet?.avatarId ?? 1, size: 40),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    petName,
                    style: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                      color: _ink,
                    ),
                  ),
                  Text(
                    pet == null
                        ? Zh.chatGreeting
                        : '${pet.mood} · Lv.${pet.level}',
                    style: const TextStyle(fontSize: 12, color: _muted),
                  ),
                ],
              ),
            ),
          ],
        ),
        actions: [
          if (state.syncing)
            const Padding(
              padding: EdgeInsets.only(right: 16),
              child: SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
            ),
        ],
      ),
      body: Column(
        children: [
          if (state.error != null) _ErrorBanner(message: state.error!),
          if (state.quotaExhausted)
            _QuotaBanner(onUpgrade: () => context.push('/subscription')),
          if (state.newMemories.isNotEmpty)
            _MemoryBanner(memories: state.newMemories),
          Expanded(
            child: _buildBody(
              state,
              petName,
              petAvatarId: petAvatarId,
              userAvatarId: userAvatarId,
            ),
          ),
          _InputBar(
            controller: _inputCtrl,
            sending: state.sending,
            onSend: _send,
          ),
        ],
      ),
    );
  }

  Widget _buildBody(
    ChatState state,
    String petName, {
    required int petAvatarId,
    required int userAvatarId,
  }) {
    if (state.loading && state.messages.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }
    if (state.messages.isEmpty) {
      return _EmptyView(
        petName: petName,
        error: state.error,
        onChip: (text) {
          _inputCtrl.text = text;
          setState(() {});
        },
        onRetry: () =>
            ref.read(chatControllerProvider(widget.petId).notifier).open(),
      );
    }
    return ListView.builder(
      controller: _scrollCtrl,
      reverse: true,
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
      itemCount: state.messages.length + (state.hasMore ? 1 : 0),
      itemBuilder: (context, index) {
        if (state.hasMore && index == state.messages.length) {
          return Padding(
            padding: const EdgeInsets.symmetric(vertical: 12),
            child: Center(
              child: TextButton(
                onPressed: state.syncing
                    ? null
                    : () => ref
                          .read(chatControllerProvider(widget.petId).notifier)
                          .sync(),
                child: const Text(Zh.chatLoadMore),
              ),
            ),
          );
        }
        final message = state.messages[state.messages.length - 1 - index];
        return _MessageBubble(
          message: message,
          onRetry: () => ref
              .read(chatControllerProvider(widget.petId).notifier)
              .retry(message.clientMsgId ?? ''),
          petAvatarId: petAvatarId,
          userAvatarId: userAvatarId,
        );
      },
    );
  }
}

class _ErrorBanner extends StatelessWidget {
  const _ErrorBanner({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      color: _danger.withValues(alpha: 0.08),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          const Icon(Icons.error_outline, size: 16, color: _danger),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              message,
              style: const TextStyle(fontSize: 12, color: _danger),
            ),
          ),
        ],
      ),
    );
  }
}

/// 当日 AI 额度耗尽引导（服务端 2501）：引导前往订阅页。
class _QuotaBanner extends StatelessWidget {
  const _QuotaBanner({required this.onUpgrade});

  final VoidCallback onUpgrade;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      color: _accent.withValues(alpha: 0.14),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          const Icon(Icons.workspace_premium_rounded, size: 16, color: _accent),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              Zh.chatQuotaExhausted,
              style: const TextStyle(fontSize: 12, color: _ink),
            ),
          ),
          TextButton(
            onPressed: onUpgrade,
            child: Text(
              Zh.goSubscribe,
              style: const TextStyle(fontSize: 12, color: _primary),
            ),
          ),
        ],
      ),
    );
  }
}

/// 新记忆形成提示（MILESTONES W3：聊天页可展示轻量记忆形成提示）。
class _MemoryBanner extends StatelessWidget {
  const _MemoryBanner({required this.memories});

  final List<PetMemory> memories;

  @override
  Widget build(BuildContext context) {
    final text = memories.map((m) => m.content).join('、');
    return Container(
      width: double.infinity,
      color: _primary.withValues(alpha: 0.08),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          const Icon(Icons.psychology_alt_rounded, size: 16, color: _primary),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              '${Zh.chatNewMemory}：$text',
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(fontSize: 12, color: _primary),
            ),
          ),
        ],
      ),
    );
  }
}

class _EmptyView extends StatelessWidget {
  const _EmptyView({
    required this.petName,
    required this.onChip,
    required this.onRetry,
    this.error,
  });

  final String petName;
  final ValueChanged<String> onChip;
  final VoidCallback onRetry;
  final String? error;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 88,
              height: 88,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: LinearGradient(
                  colors: [
                    _primary.withValues(alpha: 0.18),
                    _accent.withValues(alpha: 0.22),
                  ],
                ),
              ),
              child: const Icon(Icons.auto_awesome, color: _primary, size: 40),
            ),
            const SizedBox(height: 20),
            Text(
              Zh.chatGreeting,
              style: const TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.w600,
                color: _ink,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              Zh.chatEmptyHint,
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 13, color: _muted),
            ),
            if (error != null) ...[
              const SizedBox(height: 16),
              OutlinedButton.icon(
                onPressed: onRetry,
                icon: const Icon(Icons.refresh),
                label: const Text(Zh.retry),
              ),
            ],
            const SizedBox(height: 24),
            Wrap(
              spacing: 10,
              children: [
                for (final chip in [Zh.chatChip1, Zh.chatChip2, Zh.chatChip3])
                  ActionChip(
                    label: Text(chip),
                    backgroundColor: Colors.white,
                    side: const BorderSide(color: _primarySoft),
                    onPressed: () => onChip(chip),
                  ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _MessageBubble extends StatelessWidget {
  const _MessageBubble({
    required this.message,
    required this.onRetry,
    required this.petAvatarId,
    required this.userAvatarId,
  });

  final PetMessage message;
  final VoidCallback onRetry;

  /// 宠物侧头像（1~12）；取不到宠物时兜底 1。
  final int petAvatarId;

  /// 用户侧头像（1~8）；未登录或资料未就绪时兜底 1。
  final int userAvatarId;

  @override
  Widget build(BuildContext context) {
    final isUser = message.isUser;
    final failed = message.isFailed;
    final sending = message.sendState == PetSendState.sending;

    // 用户消息头像在右，宠物消息头像在左。
    final Widget avatar = isUser
        ? UserAvatarWidget(avatarId: userAvatarId, size: 36)
        : PetAvatarWidget(avatarId: petAvatarId, size: 36);

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        mainAxisAlignment:
            isUser ? MainAxisAlignment.end : MainAxisAlignment.start,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (!isUser) avatar,
          if (!isUser) const SizedBox(width: 8),
          Flexible(
            child: ConstrainedBox(
              constraints: BoxConstraints(
                maxWidth: MediaQuery.of(context).size.width * 0.72,
              ),
              child: Column(
                crossAxisAlignment:
                    isUser ? CrossAxisAlignment.end : CrossAxisAlignment.start,
                children: [
                  AnimatedOpacity(
                    opacity: sending ? 0.6 : 1,
                    duration: const Duration(milliseconds: 180),
                    child: DecoratedBox(
                      decoration: BoxDecoration(
                        gradient: isUser
                            ? const LinearGradient(
                                colors: [_primary, _primarySoft],
                              )
                            : null,
                        color: isUser ? null : Colors.white,
                        border: isUser
                            ? null
                            : Border.all(
                                color:
                                    failed ? _danger : const Color(0xFFEDE6FF),
                              ),
                        borderRadius: BorderRadius.circular(20).copyWith(
                          bottomRight:
                              isUser ? const Radius.circular(6) : null,
                          bottomLeft: isUser ? null : const Radius.circular(6),
                        ),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.black.withValues(alpha: 0.04),
                            blurRadius: 12,
                            offset: const Offset(0, 4),
                          ),
                        ],
                      ),
                      child: Padding(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 12,
                        ),
                        child: Text(
                          message.content,
                          style: TextStyle(
                            fontSize: 14,
                            height: 1.4,
                            color: isUser ? Colors.white : _ink,
                          ),
                        ),
                      ),
                    ),
                  ),
                  if (sending)
                    const Padding(
                      padding: EdgeInsets.only(top: 4, right: 4),
                      child: SizedBox(
                        width: 10,
                        height: 10,
                        child: CircularProgressIndicator(strokeWidth: 1.5),
                      ),
                    ),
                  if (failed)
                    Padding(
                      padding: const EdgeInsets.only(top: 4),
                      child: GestureDetector(
                        onTap: onRetry,
                        child: const Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(Icons.refresh, size: 13, color: _danger),
                            SizedBox(width: 4),
                            Text(
                              Zh.chatFailedHint,
                              style: TextStyle(fontSize: 11, color: _danger),
                            ),
                          ],
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ),
          if (isUser) const SizedBox(width: 8),
          if (isUser) avatar,
        ],
      ),
    );
  }
}

class _InputBar extends StatelessWidget {
  const _InputBar({
    required this.controller,
    required this.sending,
    required this.onSend,
  });

  final TextEditingController controller;
  final bool sending;
  final VoidCallback onSend;

  @override
  Widget build(BuildContext context) {
    final enabled = controller.text.trim().isNotEmpty && !sending;
    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 12),
        child: Row(
          children: [
            Expanded(
              child: DecoratedBox(
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(24),
                  border: Border.all(color: const Color(0xFFEDE6FF)),
                ),
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: TextField(
                    controller: controller,
                    minLines: 1,
                    maxLines: 4,
                    textInputAction: TextInputAction.send,
                    onSubmitted: (_) => enabled ? onSend() : null,
                    decoration: const InputDecoration(
                      hintText: Zh.chatInputHint,
                      hintStyle: TextStyle(fontSize: 14, color: _muted),
                      border: InputBorder.none,
                    ),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 10),
            AnimatedContainer(
              duration: const Duration(milliseconds: 180),
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: enabled
                    ? const LinearGradient(colors: [_primary, _accent])
                    : LinearGradient(
                        colors: [Colors.grey.shade300, Colors.grey.shade300],
                      ),
              ),
              child: IconButton(
                onPressed: enabled ? onSend : null,
                icon: sending
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Icon(Icons.send_rounded, color: Colors.white),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
