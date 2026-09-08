import 'package:flutter/material.dart';

/// 预置头像组件：id（1~8）→ assets/avatars/avatar_{id}.png 映射。
/// 占位图由 scripts/gen_avatars.ps1 生成；正式插画直接替换同名文件。
class AvatarWidget extends StatelessWidget {
  const AvatarWidget({super.key, required this.avatarId, this.size = 48});

  final int avatarId;
  final double size;

  static String assetOf(int id) => 'assets/avatars/avatar_$id.png';

  @override
  Widget build(BuildContext context) {
    final id = avatarId >= 1 && avatarId <= 8 ? avatarId : 1;
    return ClipOval(
      child: Image.asset(
        assetOf(id),
        width: size,
        height: size,
        fit: BoxFit.cover,
        errorBuilder: (_, _, _) => _fallback(id),
      ),
    );
  }

  Widget _fallback(int id) => Container(
        width: size,
        height: size,
        color: Colors.blueGrey,
        alignment: Alignment.center,
        child: Text(
          '$id',
          style: TextStyle(
            color: Colors.white,
            fontSize: size * 0.45,
            fontWeight: FontWeight.bold,
          ),
        ),
      );
}
