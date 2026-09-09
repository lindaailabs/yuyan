import '../model/pet_message.dart';
import '../model/pet_memory.dart';
import '../remote/api_client.dart';

/// 记忆仓库：查看与删除长期记忆（guide §3.3：UI 不直接调 Dio）。
class PetMemoryRepository {
  const PetMemoryRepository(this._api);

  final ApiClient _api;

  /// 宠物的 active 记忆列表。
  Future<List<PetMemory>> listMemories(int petId) async {
    final items = await _api.getList('/pets/$petId/memories');
    return items
        .map((e) => PetMemory.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// 软删除一条记忆（服务端 status=3，不物理删除）。
  Future<void> deleteMemory(int memoryId) async {
    await _api.delete('/pet-memories/$memoryId');
  }
}

/// 删除后从列表中移除，保持不可变更新。
List<PetMemory> removeMemoryById(List<PetMemory> source, int id) =>
    source.where((m) => m.id != id).toList();

/// 新形成的记忆（聊天页提示用）。
List<PetMemory> memoriesFromResult(SendMessageResult result) =>
    result.newMemories;
