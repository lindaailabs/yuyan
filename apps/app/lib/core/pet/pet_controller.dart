import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/model/pet.dart';
import '../../data/remote/api_exception.dart';
import '../../data/repository/pet_repository.dart';
import '../l10n/zh.dart';

class PetHomeState {
  const PetHomeState({
    this.loading = false,
    this.creating = false,
    this.pets = const [],
    this.current,
    this.error,
  });

  final bool loading;
  final bool creating;
  final List<PetProfile> pets;
  final PetProfile? current;
  final String? error;

  bool get hasPet => current != null;

  PetHomeState copyWith({
    bool? loading,
    bool? creating,
    List<PetProfile>? pets,
    PetProfile? current,
    String? error,
    bool clearError = false,
  }) => PetHomeState(
    loading: loading ?? this.loading,
    creating: creating ?? this.creating,
    pets: pets ?? this.pets,
    current: current ?? this.current,
    error: clearError ? null : error ?? this.error,
  );
}

class PetController extends StateNotifier<PetHomeState> {
  PetController(this._repo) : super(const PetHomeState());

  final PetRepository _repo;

  Future<void> load() async {
    state = state.copyWith(loading: true, clearError: true);
    try {
      final pets = await _repo.listPets();
      state = PetHomeState(
        pets: pets,
        current: pets.isEmpty ? null : pets.first,
      );
    } on ApiException catch (e) {
      state = state.copyWith(loading: false, error: e.msg);
    } catch (_) {
      state = state.copyWith(loading: false, error: Zh.errorOccurred);
    }
  }

  Future<bool> create({
    required String name,
    int avatarId = 1,
    String species = 'swallow',
  }) async {
    state = state.copyWith(creating: true, clearError: true);
    try {
      final pet = await _repo.createPet(
        name: name,
        avatarId: avatarId,
        species: species,
      );
      state = PetHomeState(pets: [pet, ...state.pets], current: pet);
      return true;
    } on ApiException catch (e) {
      state = state.copyWith(creating: false, error: e.msg);
      return false;
    } catch (_) {
      state = state.copyWith(creating: false, error: Zh.errorOccurred);
      return false;
    }
  }
}
