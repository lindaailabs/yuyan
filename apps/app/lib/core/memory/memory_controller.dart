import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/model/pet_memory.dart';
import '../../data/remote/api_exception.dart';
import '../../data/repository/pet_memory_repository.dart';
import '../l10n/zh.dart';

/// 记忆页状态：列表 + 加载/删除/失败态。
class MemoryState {
  const MemoryState({
    this.memories = const [],
    this.loading = false,
    this.deletingIds = const {},
    this.error,
  });

  final List<PetMemory> memories;
  final bool loading;

  /// 正在删除的记忆 id（按钮防抖）。
  final Set<int> deletingIds;
  final String? error;

  MemoryState copyWith({
    List<PetMemory>? memories,
    bool? loading,
    Set<int>? deletingIds,
    String? error,
    bool clearError = false,
  }) =>
      MemoryState(
        memories: memories ?? this.memories,
        loading: loading ?? this.loading,
        deletingIds: deletingIds ?? this.deletingIds,
        error: clearError ? null : error ?? this.error,
      );
}

/// 记忆控制器：加载、软删除（成功后局部移除，不整页刷新）。
class MemoryController extends StateNotifier<MemoryState> {
  MemoryController(this._repo, {required this.petId})
      : super(const MemoryState());

  final PetMemoryRepository _repo;
  final int petId;

  Future<void> load() async {
    state = state.copyWith(loading: true, clearError: true);
    try {
      final items = await _repo.listMemories(petId);
      state = MemoryState(memories: items);
    } on ApiException catch (e) {
      state = state.copyWith(loading: false, error: e.msg);
    } catch (_) {
      state = state.copyWith(loading: false, error: Zh.errorOccurred);
    }
  }

  /// 删除记忆：成功即从列表移除（服务端软删除，删除后不再召回）。
  Future<bool> delete(int memoryId) async {
    final deleting = {...state.deletingIds}..add(memoryId);
    state = state.copyWith(deletingIds: deleting, clearError: true);
    try {
      await _repo.deleteMemory(memoryId);
      final next = state.memories.where((m) => m.id != memoryId).toList();
      final remain = {...state.deletingIds}..remove(memoryId);
      state = state.copyWith(memories: next, deletingIds: remain);
      return true;
    } on ApiException catch (e) {
      _finish(memoryId, e.msg);
      return false;
    } catch (_) {
      _finish(memoryId, Zh.memoryDeleteFailed);
      return false;
    }
  }

  void _finish(int memoryId, String error) {
    final remain = {...state.deletingIds}..remove(memoryId);
    state = state.copyWith(deletingIds: remain, error: error);
  }
}
