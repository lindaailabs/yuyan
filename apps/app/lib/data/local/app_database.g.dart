// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'app_database.dart';

// ignore_for_file: type=lint
mixin _$MessageDaoMixin on DatabaseAccessor<AppDatabase> {
  $LocalMessagesTable get localMessages => attachedDatabase.localMessages;
  MessageDaoManager get managers => MessageDaoManager(this);
}

class MessageDaoManager {
  final _$MessageDaoMixin _db;
  MessageDaoManager(this._db);
  $$LocalMessagesTableTableManager get localMessages =>
      $$LocalMessagesTableTableManager(_db.attachedDatabase, _db.localMessages);
}

class $LocalMessagesTable extends LocalMessages
    with TableInfo<$LocalMessagesTable, LocalMessageRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $LocalMessagesTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _localIdMeta = const VerificationMeta(
    'localId',
  );
  @override
  late final GeneratedColumn<int> localId = GeneratedColumn<int>(
    'local_id',
    aliasedName,
    false,
    hasAutoIncrement: true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'PRIMARY KEY AUTOINCREMENT',
    ),
  );
  static const VerificationMeta _convIdMeta = const VerificationMeta('convId');
  @override
  late final GeneratedColumn<int> convId = GeneratedColumn<int>(
    'conv_id',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _serverIdMeta = const VerificationMeta(
    'serverId',
  );
  @override
  late final GeneratedColumn<int> serverId = GeneratedColumn<int>(
    'server_id',
    aliasedName,
    true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _roleMeta = const VerificationMeta('role');
  @override
  late final GeneratedColumn<String> role = GeneratedColumn<String>(
    'role',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _contentMeta = const VerificationMeta(
    'content',
  );
  @override
  late final GeneratedColumn<String> content = GeneratedColumn<String>(
    'content',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _clientMsgIdMeta = const VerificationMeta(
    'clientMsgId',
  );
  @override
  late final GeneratedColumn<String> clientMsgId = GeneratedColumn<String>(
    'client_msg_id',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _sendStateMeta = const VerificationMeta(
    'sendState',
  );
  @override
  late final GeneratedColumn<int> sendState = GeneratedColumn<int>(
    'send_state',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _errorCodeMeta = const VerificationMeta(
    'errorCode',
  );
  @override
  late final GeneratedColumn<int> errorCode = GeneratedColumn<int>(
    'error_code',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _createdAtMeta = const VerificationMeta(
    'createdAt',
  );
  @override
  late final GeneratedColumn<int> createdAt = GeneratedColumn<int>(
    'created_at',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  @override
  List<GeneratedColumn> get $columns => [
    localId,
    convId,
    serverId,
    role,
    content,
    clientMsgId,
    sendState,
    errorCode,
    createdAt,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'local_messages';
  @override
  VerificationContext validateIntegrity(
    Insertable<LocalMessageRow> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('local_id')) {
      context.handle(
        _localIdMeta,
        localId.isAcceptableOrUnknown(data['local_id']!, _localIdMeta),
      );
    }
    if (data.containsKey('conv_id')) {
      context.handle(
        _convIdMeta,
        convId.isAcceptableOrUnknown(data['conv_id']!, _convIdMeta),
      );
    } else if (isInserting) {
      context.missing(_convIdMeta);
    }
    if (data.containsKey('server_id')) {
      context.handle(
        _serverIdMeta,
        serverId.isAcceptableOrUnknown(data['server_id']!, _serverIdMeta),
      );
    }
    if (data.containsKey('role')) {
      context.handle(
        _roleMeta,
        role.isAcceptableOrUnknown(data['role']!, _roleMeta),
      );
    } else if (isInserting) {
      context.missing(_roleMeta);
    }
    if (data.containsKey('content')) {
      context.handle(
        _contentMeta,
        content.isAcceptableOrUnknown(data['content']!, _contentMeta),
      );
    } else if (isInserting) {
      context.missing(_contentMeta);
    }
    if (data.containsKey('client_msg_id')) {
      context.handle(
        _clientMsgIdMeta,
        clientMsgId.isAcceptableOrUnknown(
          data['client_msg_id']!,
          _clientMsgIdMeta,
        ),
      );
    }
    if (data.containsKey('send_state')) {
      context.handle(
        _sendStateMeta,
        sendState.isAcceptableOrUnknown(data['send_state']!, _sendStateMeta),
      );
    }
    if (data.containsKey('error_code')) {
      context.handle(
        _errorCodeMeta,
        errorCode.isAcceptableOrUnknown(data['error_code']!, _errorCodeMeta),
      );
    }
    if (data.containsKey('created_at')) {
      context.handle(
        _createdAtMeta,
        createdAt.isAcceptableOrUnknown(data['created_at']!, _createdAtMeta),
      );
    } else if (isInserting) {
      context.missing(_createdAtMeta);
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {localId};
  @override
  LocalMessageRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return LocalMessageRow(
      localId: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}local_id'],
      )!,
      convId: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}conv_id'],
      )!,
      serverId: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}server_id'],
      ),
      role: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}role'],
      )!,
      content: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}content'],
      )!,
      clientMsgId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}client_msg_id'],
      ),
      sendState: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}send_state'],
      )!,
      errorCode: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}error_code'],
      )!,
      createdAt: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}created_at'],
      )!,
    );
  }

  @override
  $LocalMessagesTable createAlias(String alias) {
    return $LocalMessagesTable(attachedDatabase, alias);
  }
}

class LocalMessageRow extends DataClass implements Insertable<LocalMessageRow> {
  final int localId;
  final int convId;
  final int? serverId;
  final String role;
  final String content;
  final String? clientMsgId;
  final int sendState;
  final int errorCode;
  final int createdAt;
  const LocalMessageRow({
    required this.localId,
    required this.convId,
    this.serverId,
    required this.role,
    required this.content,
    this.clientMsgId,
    required this.sendState,
    required this.errorCode,
    required this.createdAt,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['local_id'] = Variable<int>(localId);
    map['conv_id'] = Variable<int>(convId);
    if (!nullToAbsent || serverId != null) {
      map['server_id'] = Variable<int>(serverId);
    }
    map['role'] = Variable<String>(role);
    map['content'] = Variable<String>(content);
    if (!nullToAbsent || clientMsgId != null) {
      map['client_msg_id'] = Variable<String>(clientMsgId);
    }
    map['send_state'] = Variable<int>(sendState);
    map['error_code'] = Variable<int>(errorCode);
    map['created_at'] = Variable<int>(createdAt);
    return map;
  }

  LocalMessagesCompanion toCompanion(bool nullToAbsent) {
    return LocalMessagesCompanion(
      localId: Value(localId),
      convId: Value(convId),
      serverId: serverId == null && nullToAbsent
          ? const Value.absent()
          : Value(serverId),
      role: Value(role),
      content: Value(content),
      clientMsgId: clientMsgId == null && nullToAbsent
          ? const Value.absent()
          : Value(clientMsgId),
      sendState: Value(sendState),
      errorCode: Value(errorCode),
      createdAt: Value(createdAt),
    );
  }

  factory LocalMessageRow.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return LocalMessageRow(
      localId: serializer.fromJson<int>(json['localId']),
      convId: serializer.fromJson<int>(json['convId']),
      serverId: serializer.fromJson<int?>(json['serverId']),
      role: serializer.fromJson<String>(json['role']),
      content: serializer.fromJson<String>(json['content']),
      clientMsgId: serializer.fromJson<String?>(json['clientMsgId']),
      sendState: serializer.fromJson<int>(json['sendState']),
      errorCode: serializer.fromJson<int>(json['errorCode']),
      createdAt: serializer.fromJson<int>(json['createdAt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'localId': serializer.toJson<int>(localId),
      'convId': serializer.toJson<int>(convId),
      'serverId': serializer.toJson<int?>(serverId),
      'role': serializer.toJson<String>(role),
      'content': serializer.toJson<String>(content),
      'clientMsgId': serializer.toJson<String?>(clientMsgId),
      'sendState': serializer.toJson<int>(sendState),
      'errorCode': serializer.toJson<int>(errorCode),
      'createdAt': serializer.toJson<int>(createdAt),
    };
  }

  LocalMessageRow copyWith({
    int? localId,
    int? convId,
    Value<int?> serverId = const Value.absent(),
    String? role,
    String? content,
    Value<String?> clientMsgId = const Value.absent(),
    int? sendState,
    int? errorCode,
    int? createdAt,
  }) => LocalMessageRow(
    localId: localId ?? this.localId,
    convId: convId ?? this.convId,
    serverId: serverId.present ? serverId.value : this.serverId,
    role: role ?? this.role,
    content: content ?? this.content,
    clientMsgId: clientMsgId.present ? clientMsgId.value : this.clientMsgId,
    sendState: sendState ?? this.sendState,
    errorCode: errorCode ?? this.errorCode,
    createdAt: createdAt ?? this.createdAt,
  );
  LocalMessageRow copyWithCompanion(LocalMessagesCompanion data) {
    return LocalMessageRow(
      localId: data.localId.present ? data.localId.value : this.localId,
      convId: data.convId.present ? data.convId.value : this.convId,
      serverId: data.serverId.present ? data.serverId.value : this.serverId,
      role: data.role.present ? data.role.value : this.role,
      content: data.content.present ? data.content.value : this.content,
      clientMsgId: data.clientMsgId.present
          ? data.clientMsgId.value
          : this.clientMsgId,
      sendState: data.sendState.present ? data.sendState.value : this.sendState,
      errorCode: data.errorCode.present ? data.errorCode.value : this.errorCode,
      createdAt: data.createdAt.present ? data.createdAt.value : this.createdAt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('LocalMessageRow(')
          ..write('localId: $localId, ')
          ..write('convId: $convId, ')
          ..write('serverId: $serverId, ')
          ..write('role: $role, ')
          ..write('content: $content, ')
          ..write('clientMsgId: $clientMsgId, ')
          ..write('sendState: $sendState, ')
          ..write('errorCode: $errorCode, ')
          ..write('createdAt: $createdAt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
    localId,
    convId,
    serverId,
    role,
    content,
    clientMsgId,
    sendState,
    errorCode,
    createdAt,
  );
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is LocalMessageRow &&
          other.localId == this.localId &&
          other.convId == this.convId &&
          other.serverId == this.serverId &&
          other.role == this.role &&
          other.content == this.content &&
          other.clientMsgId == this.clientMsgId &&
          other.sendState == this.sendState &&
          other.errorCode == this.errorCode &&
          other.createdAt == this.createdAt);
}

class LocalMessagesCompanion extends UpdateCompanion<LocalMessageRow> {
  final Value<int> localId;
  final Value<int> convId;
  final Value<int?> serverId;
  final Value<String> role;
  final Value<String> content;
  final Value<String?> clientMsgId;
  final Value<int> sendState;
  final Value<int> errorCode;
  final Value<int> createdAt;
  const LocalMessagesCompanion({
    this.localId = const Value.absent(),
    this.convId = const Value.absent(),
    this.serverId = const Value.absent(),
    this.role = const Value.absent(),
    this.content = const Value.absent(),
    this.clientMsgId = const Value.absent(),
    this.sendState = const Value.absent(),
    this.errorCode = const Value.absent(),
    this.createdAt = const Value.absent(),
  });
  LocalMessagesCompanion.insert({
    this.localId = const Value.absent(),
    required int convId,
    this.serverId = const Value.absent(),
    required String role,
    required String content,
    this.clientMsgId = const Value.absent(),
    this.sendState = const Value.absent(),
    this.errorCode = const Value.absent(),
    required int createdAt,
  }) : convId = Value(convId),
       role = Value(role),
       content = Value(content),
       createdAt = Value(createdAt);
  static Insertable<LocalMessageRow> custom({
    Expression<int>? localId,
    Expression<int>? convId,
    Expression<int>? serverId,
    Expression<String>? role,
    Expression<String>? content,
    Expression<String>? clientMsgId,
    Expression<int>? sendState,
    Expression<int>? errorCode,
    Expression<int>? createdAt,
  }) {
    return RawValuesInsertable({
      if (localId != null) 'local_id': localId,
      if (convId != null) 'conv_id': convId,
      if (serverId != null) 'server_id': serverId,
      if (role != null) 'role': role,
      if (content != null) 'content': content,
      if (clientMsgId != null) 'client_msg_id': clientMsgId,
      if (sendState != null) 'send_state': sendState,
      if (errorCode != null) 'error_code': errorCode,
      if (createdAt != null) 'created_at': createdAt,
    });
  }

  LocalMessagesCompanion copyWith({
    Value<int>? localId,
    Value<int>? convId,
    Value<int?>? serverId,
    Value<String>? role,
    Value<String>? content,
    Value<String?>? clientMsgId,
    Value<int>? sendState,
    Value<int>? errorCode,
    Value<int>? createdAt,
  }) {
    return LocalMessagesCompanion(
      localId: localId ?? this.localId,
      convId: convId ?? this.convId,
      serverId: serverId ?? this.serverId,
      role: role ?? this.role,
      content: content ?? this.content,
      clientMsgId: clientMsgId ?? this.clientMsgId,
      sendState: sendState ?? this.sendState,
      errorCode: errorCode ?? this.errorCode,
      createdAt: createdAt ?? this.createdAt,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (localId.present) {
      map['local_id'] = Variable<int>(localId.value);
    }
    if (convId.present) {
      map['conv_id'] = Variable<int>(convId.value);
    }
    if (serverId.present) {
      map['server_id'] = Variable<int>(serverId.value);
    }
    if (role.present) {
      map['role'] = Variable<String>(role.value);
    }
    if (content.present) {
      map['content'] = Variable<String>(content.value);
    }
    if (clientMsgId.present) {
      map['client_msg_id'] = Variable<String>(clientMsgId.value);
    }
    if (sendState.present) {
      map['send_state'] = Variable<int>(sendState.value);
    }
    if (errorCode.present) {
      map['error_code'] = Variable<int>(errorCode.value);
    }
    if (createdAt.present) {
      map['created_at'] = Variable<int>(createdAt.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('LocalMessagesCompanion(')
          ..write('localId: $localId, ')
          ..write('convId: $convId, ')
          ..write('serverId: $serverId, ')
          ..write('role: $role, ')
          ..write('content: $content, ')
          ..write('clientMsgId: $clientMsgId, ')
          ..write('sendState: $sendState, ')
          ..write('errorCode: $errorCode, ')
          ..write('createdAt: $createdAt')
          ..write(')'))
        .toString();
  }
}

abstract class _$AppDatabase extends GeneratedDatabase {
  _$AppDatabase(QueryExecutor e) : super(e);
  $AppDatabaseManager get managers => $AppDatabaseManager(this);
  late final $LocalMessagesTable localMessages = $LocalMessagesTable(this);
  late final MessageDao messageDao = MessageDao(this as AppDatabase);
  @override
  Iterable<TableInfo<Table, Object?>> get allTables =>
      allSchemaEntities.whereType<TableInfo<Table, Object?>>();
  @override
  List<DatabaseSchemaEntity> get allSchemaEntities => [localMessages];
}

typedef $$LocalMessagesTableCreateCompanionBuilder =
    LocalMessagesCompanion Function({
      Value<int> localId,
      required int convId,
      Value<int?> serverId,
      required String role,
      required String content,
      Value<String?> clientMsgId,
      Value<int> sendState,
      Value<int> errorCode,
      required int createdAt,
    });
typedef $$LocalMessagesTableUpdateCompanionBuilder =
    LocalMessagesCompanion Function({
      Value<int> localId,
      Value<int> convId,
      Value<int?> serverId,
      Value<String> role,
      Value<String> content,
      Value<String?> clientMsgId,
      Value<int> sendState,
      Value<int> errorCode,
      Value<int> createdAt,
    });

class $$LocalMessagesTableFilterComposer
    extends Composer<_$AppDatabase, $LocalMessagesTable> {
  $$LocalMessagesTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<int> get localId => $composableBuilder(
    column: $table.localId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get convId => $composableBuilder(
    column: $table.convId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get serverId => $composableBuilder(
    column: $table.serverId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get role => $composableBuilder(
    column: $table.role,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get content => $composableBuilder(
    column: $table.content,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get clientMsgId => $composableBuilder(
    column: $table.clientMsgId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get sendState => $composableBuilder(
    column: $table.sendState,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get errorCode => $composableBuilder(
    column: $table.errorCode,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get createdAt => $composableBuilder(
    column: $table.createdAt,
    builder: (column) => ColumnFilters(column),
  );
}

class $$LocalMessagesTableOrderingComposer
    extends Composer<_$AppDatabase, $LocalMessagesTable> {
  $$LocalMessagesTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<int> get localId => $composableBuilder(
    column: $table.localId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get convId => $composableBuilder(
    column: $table.convId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get serverId => $composableBuilder(
    column: $table.serverId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get role => $composableBuilder(
    column: $table.role,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get content => $composableBuilder(
    column: $table.content,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get clientMsgId => $composableBuilder(
    column: $table.clientMsgId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get sendState => $composableBuilder(
    column: $table.sendState,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get errorCode => $composableBuilder(
    column: $table.errorCode,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get createdAt => $composableBuilder(
    column: $table.createdAt,
    builder: (column) => ColumnOrderings(column),
  );
}

class $$LocalMessagesTableAnnotationComposer
    extends Composer<_$AppDatabase, $LocalMessagesTable> {
  $$LocalMessagesTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<int> get localId =>
      $composableBuilder(column: $table.localId, builder: (column) => column);

  GeneratedColumn<int> get convId =>
      $composableBuilder(column: $table.convId, builder: (column) => column);

  GeneratedColumn<int> get serverId =>
      $composableBuilder(column: $table.serverId, builder: (column) => column);

  GeneratedColumn<String> get role =>
      $composableBuilder(column: $table.role, builder: (column) => column);

  GeneratedColumn<String> get content =>
      $composableBuilder(column: $table.content, builder: (column) => column);

  GeneratedColumn<String> get clientMsgId => $composableBuilder(
    column: $table.clientMsgId,
    builder: (column) => column,
  );

  GeneratedColumn<int> get sendState =>
      $composableBuilder(column: $table.sendState, builder: (column) => column);

  GeneratedColumn<int> get errorCode =>
      $composableBuilder(column: $table.errorCode, builder: (column) => column);

  GeneratedColumn<int> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => column);
}

class $$LocalMessagesTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $LocalMessagesTable,
          LocalMessageRow,
          $$LocalMessagesTableFilterComposer,
          $$LocalMessagesTableOrderingComposer,
          $$LocalMessagesTableAnnotationComposer,
          $$LocalMessagesTableCreateCompanionBuilder,
          $$LocalMessagesTableUpdateCompanionBuilder,
          (
            LocalMessageRow,
            BaseReferences<_$AppDatabase, $LocalMessagesTable, LocalMessageRow>,
          ),
          LocalMessageRow,
          PrefetchHooks Function()
        > {
  $$LocalMessagesTableTableManager(_$AppDatabase db, $LocalMessagesTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$LocalMessagesTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$LocalMessagesTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$LocalMessagesTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<int> localId = const Value.absent(),
                Value<int> convId = const Value.absent(),
                Value<int?> serverId = const Value.absent(),
                Value<String> role = const Value.absent(),
                Value<String> content = const Value.absent(),
                Value<String?> clientMsgId = const Value.absent(),
                Value<int> sendState = const Value.absent(),
                Value<int> errorCode = const Value.absent(),
                Value<int> createdAt = const Value.absent(),
              }) => LocalMessagesCompanion(
                localId: localId,
                convId: convId,
                serverId: serverId,
                role: role,
                content: content,
                clientMsgId: clientMsgId,
                sendState: sendState,
                errorCode: errorCode,
                createdAt: createdAt,
              ),
          createCompanionCallback:
              ({
                Value<int> localId = const Value.absent(),
                required int convId,
                Value<int?> serverId = const Value.absent(),
                required String role,
                required String content,
                Value<String?> clientMsgId = const Value.absent(),
                Value<int> sendState = const Value.absent(),
                Value<int> errorCode = const Value.absent(),
                required int createdAt,
              }) => LocalMessagesCompanion.insert(
                localId: localId,
                convId: convId,
                serverId: serverId,
                role: role,
                content: content,
                clientMsgId: clientMsgId,
                sendState: sendState,
                errorCode: errorCode,
                createdAt: createdAt,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable<$LocalMessagesTable, LocalMessageRow>(table),
                  BaseReferences<
                    _$AppDatabase,
                    $LocalMessagesTable,
                    LocalMessageRow
                  >(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$LocalMessagesTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $LocalMessagesTable,
      LocalMessageRow,
      $$LocalMessagesTableFilterComposer,
      $$LocalMessagesTableOrderingComposer,
      $$LocalMessagesTableAnnotationComposer,
      $$LocalMessagesTableCreateCompanionBuilder,
      $$LocalMessagesTableUpdateCompanionBuilder,
      (
        LocalMessageRow,
        BaseReferences<_$AppDatabase, $LocalMessagesTable, LocalMessageRow>,
      ),
      LocalMessageRow,
      PrefetchHooks Function()
    >;

class $AppDatabaseManager {
  final _$AppDatabase _db;
  $AppDatabaseManager(this._db);
  $$LocalMessagesTableTableManager get localMessages =>
      $$LocalMessagesTableTableManager(_db, _db.localMessages);
}
