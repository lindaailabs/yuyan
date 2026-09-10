import 'package:flutter/material.dart';

/// 用户预置头像：1~8 → assets/user_avatars/user_avatar_{id}.png。
class UserAvatarWidget extends StatelessWidget {
  const UserAvatarWidget({super.key, required this.avatarId, this.size = 48});

  static const count = 8;

  final int avatarId;
  final double size;

  static String assetOf(int id) => 'assets/user_avatars/user_avatar_$id.png';

  @override
  Widget build(BuildContext context) {
    final id = avatarId >= 1 && avatarId <= count ? avatarId : 1;
    return _PresetAvatarImage(
      assetPath: assetOf(id),
      size: size,
      fallbackIcon: Icons.person,
      fallbackColor: Theme.of(context).colorScheme.primaryContainer,
    );
  }
}

/// 宠物预置头像：1~12 → assets/pet_avatars/pet_avatar_{id}.png。
class PetAvatarWidget extends StatelessWidget {
  const PetAvatarWidget({super.key, required this.avatarId, this.size = 48});

  static const count = 12;

  final int avatarId;
  final double size;

  static String assetOf(int id) => 'assets/pet_avatars/pet_avatar_$id.png';

  @override
  Widget build(BuildContext context) {
    final id = avatarId >= 1 && avatarId <= count ? avatarId : 1;
    return _PresetAvatarImage(
      assetPath: assetOf(id),
      size: size,
      fallbackIcon: Icons.pets,
      fallbackColor: Theme.of(context).colorScheme.tertiaryContainer,
    );
  }
}

/// 兼容旧调用；新代码应按场景使用 UserAvatarWidget 或 PetAvatarWidget。
class AvatarWidget extends UserAvatarWidget {
  const AvatarWidget({super.key, required super.avatarId, super.size});
}

class _PresetAvatarImage extends StatelessWidget {
  const _PresetAvatarImage({
    required this.assetPath,
    required this.size,
    required this.fallbackIcon,
    required this.fallbackColor,
  });

  final String assetPath;
  final double size;
  final IconData fallbackIcon;
  final Color fallbackColor;

  @override
  Widget build(BuildContext context) {
    return ClipOval(
      child: Image.asset(
        assetPath,
        width: size,
        height: size,
        fit: BoxFit.cover,
        errorBuilder: (_, _, _) => Container(
          width: size,
          height: size,
          color: fallbackColor,
          alignment: Alignment.center,
          child: Icon(fallbackIcon, size: size * 0.5),
        ),
      ),
    );
  }
}
