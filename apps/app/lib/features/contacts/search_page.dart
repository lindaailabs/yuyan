import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/providers.dart';
import '../../core/l10n/zh.dart';
import '../../data/model/user_profile.dart';
import '../../data/remote/api_exception.dart';
import '../shared/avatar_widget.dart';

/// 搜索页：按手机号精确搜索（脱敏展示）+ 发起好友申请（W3 启用）。
class SearchPage extends ConsumerStatefulWidget {
  const SearchPage({super.key});

  @override
  ConsumerState<SearchPage> createState() => _SearchPageState();
}

class _SearchPageState extends ConsumerState<SearchPage> {
  final _phoneCtrl = TextEditingController();
  bool _loading = false;
  List<UserSearchItem>? _results;

  /// 已成功发出申请的用户 id（按钮转「已申请」禁用态）。
  final Set<int> _sentIds = {};

  /// 正在提交申请的用户 id（防抖）。
  int? _sendingId;

  @override
  void dispose() {
    _phoneCtrl.dispose();
    super.dispose();
  }

  Future<void> _search() async {
    final phone = _phoneCtrl.text.trim();
    if (phone.isEmpty) return;
    setState(() {
      _loading = true;
      _results = null;
    });
    try {
      final items = await ref.read(authRepositoryProvider).search(phone);
      setState(() => _results = items);
    } on ApiException catch (e) {
      _showError(e.msg);
    } catch (_) {
      _showError(Zh.errorOccurred);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  /// 发起好友申请：成功→「已申请」；业务错误（2101~2107）snackbar 展示服务端 msg。
  Future<void> _sendRequest(UserSearchItem item) async {
    if (_sendingId != null || _sentIds.contains(item.id)) return;
    setState(() => _sendingId = item.id);
    try {
      await ref.read(contactsControllerProvider.notifier).sendRequest(item.id);
      if (mounted) {
        setState(() => _sentIds.add(item.id));
      }
    } on ApiException catch (e) {
      _showError(e.msg);
    } catch (_) {
      _showError(Zh.errorOccurred);
    } finally {
      if (mounted) setState(() => _sendingId = null);
    }
  }

  void _showError(String msg) {
    if (!mounted) return;
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text(Zh.searchTitle)),
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _phoneCtrl,
                    keyboardType: TextInputType.phone,
                    maxLength: 11,
                    decoration: InputDecoration(
                      labelText: Zh.searchHint,
                      counterText: '',
                      border: const OutlineInputBorder(),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                OutlinedButton(
                  onPressed: _loading ? null : _search,
                  child: Text(Zh.searchSubmit),
                ),
              ],
            ),
            const SizedBox(height: 24),
            if (_loading) const Center(child: CircularProgressIndicator()),
            if (_results != null && _results!.isEmpty)
              Center(child: Text(Zh.searchEmpty)),
            if (_results != null)
              for (final item in _results!) _resultCard(item),
          ],
        ),
      ),
    );
  }

  Widget _resultCard(UserSearchItem item) {
    final sent = _sentIds.contains(item.id);
    final sending = _sendingId == item.id;
    return Card(
      child: ListTile(
        leading: UserAvatarWidget(avatarId: item.avatarId),
        title: Text(item.nickname ?? item.phone),
        subtitle: Text(item.phone),
        trailing: FilledButton.tonal(
          onPressed: (sent || sending) ? null : () => _sendRequest(item),
          child: sending
              ? const SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : Text(sent ? Zh.searchRequestSent : Zh.searchAddFriend),
        ),
      ),
    );
  }
}
