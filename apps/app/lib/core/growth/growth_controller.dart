import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/model/pet_growth.dart';
import '../../data/remote/api_exception.dart';
import '../../data/repository/pet_growth_repository.dart';
import '../l10n/zh.dart';

/// 成长时间线状态。
class GrowthState {
  const GrowthState({
    this.events = const [],
    this.loading = false,
    this.error,
  });

  final List<GrowthEvent> events;
  final bool loading;
  final String? error;

  GrowthState copyWith({
    List<GrowthEvent>? events,
    bool? loading,
    String? error,
    bool clearError = false,
  }) =>
      GrowthState(
        events: events ?? this.events,
        loading: loading ?? this.loading,
        error: clearError ? null : error ?? this.error,
      );
}

/// 成长控制器：加载成长事件时间线。
class GrowthController extends StateNotifier<GrowthState> {
  GrowthController(this._repo, {required this.petId})
      : super(const GrowthState());

  final PetGrowthRepository _repo;
  final int petId;

  Future<void> load() async {
    state = state.copyWith(loading: true, clearError: true);
    try {
      final items = await _repo.listEvents(petId);
      state = GrowthState(events: items);
    } on ApiException catch (e) {
      state = state.copyWith(loading: false, error: e.msg);
    } catch (_) {
      state = state.copyWith(loading: false, error: Zh.errorOccurred);
    }
  }
}
