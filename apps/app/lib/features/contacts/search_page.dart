import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/providers.dart';
import '../../core/l10n/zh.dart';
import '../../data/model/user_profile.dart';
import '../../data/remote/api_exception.dart';
import '../shared/avatar_widget.dart';

/// 搜索页：按手机号精确搜索（脱敏展示）。
/// "加好友"按钮 W3 启用，此处禁用占位。
class SearchPage extends ConsumerStatefulWidget {
  const SearchPage({super.key});

  @override
  ConsumerState<SearchPage> createState() => _SearchPageState();
}

class _SearchPageState extends ConsumerState<SearchPage> {
  final _phoneCtrl = TextEditingController();
  bool _loading = false;
  List<UserSearchItem>? _results;

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
    return Card(
      child: ListTile(
        leading: AvatarWidget(avatarId: item.avatarId),
        title: Text(item.nickname ?? item.phone),
        subtitle: Text(item.phone),
        trailing: FilledButton.tonal(
          onPressed: null, // W3 接入好友功能后启用
          child: const Text(Zh.searchAddFriend),
        ),
      ),
    );
  }
}
