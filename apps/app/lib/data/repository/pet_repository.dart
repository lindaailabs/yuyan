import '../model/pet.dart';
import '../remote/api_client.dart';

/// 宠物档案仓库：UI 只依赖 repository，不触碰 remote 细节。
class PetRepository {
  const PetRepository(this._api);

  final ApiClient _api;

  Future<List<PetProfile>> listPets() async {
    final items = await _api.getList('/pets');
    return items
        .map((e) => PetProfile.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<PetProfile> createPet({
    required String name,
    String? species,
    int? avatarId,
    String? persona,
  }) async {
    final body = <String, dynamic>{'name': name};
    if (species != null) {
      body['species'] = species;
    }
    if (avatarId != null) {
      body['avatar_id'] = avatarId;
    }
    if (persona != null) {
      body['persona'] = persona;
    }
    final data = await _api.post('/pets', body: body);
    return PetProfile.fromJson(data);
  }

  Future<PetProfile> updatePet(
    int id, {
    String? name,
    String? species,
    int? avatarId,
    String? persona,
  }) async {
    final body = <String, dynamic>{};
    if (name != null) {
      body['name'] = name;
    }
    if (species != null) {
      body['species'] = species;
    }
    if (avatarId != null) {
      body['avatar_id'] = avatarId;
    }
    if (persona != null) {
      body['persona'] = persona;
    }
    final data = await _api.put('/pets/$id', body: body);
    return PetProfile.fromJson(data);
  }

  Future<PetStateView> getState(int id) async {
    final data = await _api.get('/pets/$id/state');
    return PetStateView.fromJson(data);
  }
}
