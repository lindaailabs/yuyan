import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/l10n/zh.dart';
import '../../core/providers.dart';
import '../../data/model/pet.dart';
import 'avatar_widget.dart';

const _primary = Color(0xFF7C6BF5);
const _ink = Color(0xFF2B2B34);
const _muted = Color(0xFF7A7A88);

/// 宠物切换：底部弹层列出已领养的全部宠物，点击即切换当前宠物。
///
/// [onSelected] 可选：切换完成后回调新宠物 id，供调用方做额外动作
/// （如对话页跳转到对应会话）。首页无需跳转时可不传。
void showPetSwitcherSheet(
  BuildContext context,
  WidgetRef ref, {
  ValueChanged<int>? onSelected,
}) {
  final petState = ref.read(petControllerProvider);
  final pets = petState.pets;
  final currentId = petState.current?.id;

  showModalBottomSheet<void>(
    context: context,
    shape: const RoundedRectangleBorder(
      borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
    ),
    builder: (sheetContext) => SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              Zh.petSwitchTitle,
              style: const TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w600,
                color: _ink,
              ),
            ),
            const SizedBox(height: 8),
            for (final pet in pets)
              ListTile(
                contentPadding: EdgeInsets.zero,
                leading: PetAvatarWidget(avatarId: pet.avatarId, size: 40),
                title: Text(
                  pet.name,
                  style: const TextStyle(fontSize: 14, color: _ink),
                ),
                subtitle: Text(
                  '${pet.mood} · Lv.${pet.level}',
                  style: const TextStyle(fontSize: 12, color: _muted),
                ),
                trailing: pet.id == currentId
                    ? const Icon(Icons.check_rounded, color: _primary)
                    : null,
                onTap: () {
                  ref.read(petControllerProvider.notifier).selectPet(pet.id);
                  Navigator.of(sheetContext).pop();
                  onSelected?.call(pet.id);
                },
              ),
          ],
        ),
      ),
    ),
  );
}

/// 首页横向宠物头像条：领养了多只时展示，点击直接切换。
class PetSwitchBar extends StatelessWidget {
  const PetSwitchBar({
    super.key,
    required this.pets,
    required this.currentId,
    required this.onSelect,
  });

  final List<PetProfile> pets;
  final int currentId;
  final ValueChanged<int> onSelect;

  @override
  Widget build(BuildContext context) {
    // 高度自适应（横向滚动 + Row），避免写死高度在字体放大时裁切内容。
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        children: [
          for (var i = 0; i < pets.length; i++) ...[
            if (i > 0) const SizedBox(width: 14),
            _PetSwitchItem(
              pet: pets[i],
              selected: pets[i].id == currentId,
              onTap: () => onSelect(pets[i].id),
            ),
          ],
        ],
      ),
    );
  }
}

class _PetSwitchItem extends StatelessWidget {
  const _PetSwitchItem({
    required this.pet,
    required this.selected,
    required this.onTap,
  });

  final PetProfile pet;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: SizedBox(
        width: 60,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.all(2),
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                border: Border.all(
                  color: selected ? _primary : Colors.transparent,
                  width: 2,
                ),
              ),
              child: PetAvatarWidget(avatarId: pet.avatarId, size: 48),
            ),
            const SizedBox(height: 4),
            Text(
              pet.name,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 12,
                color: selected ? _primary : _muted,
                fontWeight: selected ? FontWeight.w600 : FontWeight.normal,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
