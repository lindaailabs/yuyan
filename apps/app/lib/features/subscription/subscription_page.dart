import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/entitlement/entitlement_controller.dart';
import '../../core/l10n/zh.dart';
import '../../core/providers.dart';
import '../../data/model/entitlement.dart';

/// 订阅 / 权益页：展示当前套餐与额度，沙盒开通（开发环境）。
class SubscriptionPage extends ConsumerStatefulWidget {
  const SubscriptionPage({super.key});

  @override
  ConsumerState<SubscriptionPage> createState() => _SubscriptionPageState();
}

class _SubscriptionPageState extends ConsumerState<SubscriptionPage> {
  @override
  void initState() {
    super.initState();
    Future.microtask(
      () => ref.read(entitlementControllerProvider.notifier).load(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(entitlementControllerProvider);
    return Scaffold(
      backgroundColor: _bg,
      appBar: AppBar(
        backgroundColor: _bg,
        surfaceTintColor: Colors.transparent,
        title: const Text(Zh.subscription),
      ),
      body: _buildBody(state),
    );
  }

  Widget _buildBody(EntitlementState state) {
    if (state.loading && state.entitlement == null) {
      return const Center(child: CircularProgressIndicator());
    }
    final view = state.entitlement;
    return RefreshIndicator(
      onRefresh: () => ref.read(entitlementControllerProvider.notifier).load(),
      child: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          if (state.error != null)
            _Tip(message: state.error!, danger: true),
          if (view != null) ...[
            _PlanCard(view: view),
            const SizedBox(height: 16),
            _QuotaCard(view: view),
            const SizedBox(height: 20),
          ],
          _SandboxPanel(
            purchasing: state.purchasing,
            isPro: view?.isPro ?? false,
            onPurchase: (plan) =>
                ref.read(entitlementControllerProvider.notifier).purchase(plan),
          ),
          const SizedBox(height: 12),
          _Tip(message: Zh.subscriptionSandboxHint),
        ],
      ),
    );
  }
}

class _PlanCard extends StatelessWidget {
  const _PlanCard({required this.view});

  final EntitlementView view;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: view.isPro
              ? [const Color(0xFF7C6BF5), const Color(0xFF9C8CFF)]
              : [const Color(0xFFFFF9F5), const Color(0xFFF1E9FF)],
        ),
        borderRadius: BorderRadius.circular(20),
        boxShadow: const [
          BoxShadow(color: Color(0x11000000), blurRadius: 16, offset: Offset(0, 6)),
        ],
      ),
      padding: const EdgeInsets.all(20),
      child: Row(
        children: [
          Icon(
            view.isPro ? Icons.workspace_premium_rounded : Icons.pets_rounded,
            size: 36,
            color: view.isPro ? Colors.white : const Color(0xFF7C6BF5),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  view.isPro ? Zh.planPro : Zh.planFree,
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.w700,
                    color: view.isPro ? Colors.white : const Color(0xFF2B2B34),
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  view.isPro ? Zh.planProSub : Zh.planFreeSub,
                  style: TextStyle(
                    fontSize: 12,
                    color: view.isPro
                        ? Colors.white.withValues(alpha: 0.85)
                        : const Color(0xFF7A7A88),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _QuotaCard extends StatelessWidget {
  const _QuotaCard({required this.view});

  final EntitlementView view;

  @override
  Widget build(BuildContext context) {
    final q = view.quota;
    final ratio = q.dailyMessages <= 0
        ? 0.0
        : (q.dailyUsed / q.dailyMessages).clamp(0.0, 1.0);
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFEDE6FF)),
      ),
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(Zh.quotaDaily, style: _label),
              Text(
                '${q.dailyUsed} / ${q.dailyMessages}',
                style: _value,
              ),
            ],
          ),
          const SizedBox(height: 10),
          LinearProgressIndicator(
            value: ratio,
            backgroundColor: const Color(0xFFF0ECFF),
            color: const Color(0xFF7C6BF5),
            minHeight: 8,
            borderRadius: BorderRadius.circular(4),
          ),
          const SizedBox(height: 6),
          Text(
            '${Zh.quotaDailyRemain} ${q.dailyRemain}',
            style: const TextStyle(fontSize: 12, color: Color(0xFF7A7A88)),
          ),
          const Divider(height: 22, color: Color(0xFFEDE6FF)),
          Row(
            children: [
              Expanded(
                child: _Mini(
                  label: Zh.quotaMemory,
                  value: '${q.memoryLimit}',
                ),
              ),
              Expanded(
                child: _Mini(
                  label: Zh.quotaAdvancedModel,
                  value: q.advancedModel ? Zh.enabled : Zh.disabled,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _Mini extends StatelessWidget {
  const _Mini({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) => Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: const TextStyle(fontSize: 12, color: Color(0xFF7A7A88))),
          const SizedBox(height: 4),
          Text(value, style: _value),
        ],
      );
}

class _SandboxPanel extends StatelessWidget {
  const _SandboxPanel({
    required this.purchasing,
    required this.isPro,
    required this.onPurchase,
  });

  final bool purchasing;
  final bool isPro;
  final ValueChanged<String> onPurchase;

  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: const Color(0xFFEDE6FF)),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(Zh.subscriptionSandbox, style: _label),
            const SizedBox(height: 12),
            FilledButton(
              onPressed: purchasing ? null : () => onPurchase(Plan.pro),
              child: purchasing
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                    )
                  : Text(isPro ? Zh.subscriptionRenewPro : Zh.subscriptionUpgradePro),
            ),
            const SizedBox(height: 10),
            OutlinedButton(
              onPressed: purchasing ? null : () => onPurchase(Plan.free),
              child: Text(Zh.subscriptionBackFree),
            ),
          ],
        ),
      );
}

class _Tip extends StatelessWidget {
  const _Tip({required this.message, this.danger = false});

  final String message;
  final bool danger;

  @override
  Widget build(BuildContext context) => Container(
        width: double.infinity,
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
          color: (danger ? const Color(0xFFE5484D) : const Color(0xFF7C6BF5))
              .withValues(alpha: 0.08),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Text(
          message,
          style: TextStyle(
            fontSize: 12,
            color: danger ? const Color(0xFFE5484D) : const Color(0xFF7C6BF5),
          ),
        ),
      );
}

const _bg = Color(0xFFFFF9F5);
const _label = TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Color(0xFF2B2B34));
const _value = TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: Color(0xFF2B2B34));
