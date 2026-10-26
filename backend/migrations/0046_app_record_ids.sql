-- Application-local numbers supplement existing primary keys and SDK IDs.
CREATE TABLE {app_record_ids} (
  appid BIGINT NOT NULL,
  entity VARCHAR(32) NOT NULL,
  record_id BIGINT NOT NULL,
  local_id BIGINT NOT NULL,
  PRIMARY KEY (appid, entity, record_id),
  UNIQUE KEY uq_app_record_number (appid, entity, local_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE {app_record_counters} (
  appid BIGINT NOT NULL,
  entity VARCHAR(32) NOT NULL,
  last_id BIGINT NOT NULL,
  PRIMARY KEY (appid, entity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO {app_record_ids} (appid, entity, record_id, local_id)
SELECT appid, 'users', id, ROW_NUMBER() OVER (PARTITION BY appid ORDER BY id)
FROM {users} WHERE appid > 0;

INSERT INTO {app_record_ids} (appid, entity, record_id, local_id)
SELECT appid, 'user_groups', id, ROW_NUMBER() OVER (PARTITION BY appid ORDER BY id)
FROM {user_groups} WHERE appid > 0;

INSERT INTO {app_record_ids} (appid, entity, record_id, local_id)
SELECT appid, 'kamis', id, ROW_NUMBER() OVER (PARTITION BY appid ORDER BY id)
FROM {kamis} WHERE appid > 0;

INSERT INTO {app_record_ids} (appid, entity, record_id, local_id)
SELECT appid, 'kami_groups', id, ROW_NUMBER() OVER (PARTITION BY appid ORDER BY id)
FROM {kami_groups} WHERE appid > 0;

INSERT INTO {app_record_ids} (appid, entity, record_id, local_id)
SELECT appid, 'versions', id, ROW_NUMBER() OVER (PARTITION BY appid ORDER BY id)
FROM {versions} WHERE appid > 0;

INSERT INTO {app_record_ids} (appid, entity, record_id, local_id)
SELECT appid, 'notices', id, ROW_NUMBER() OVER (PARTITION BY appid ORDER BY id)
FROM {notices} WHERE appid > 0;

INSERT INTO {app_record_ids} (appid, entity, record_id, local_id)
SELECT appid, 'extends', id, ROW_NUMBER() OVER (PARTITION BY appid ORDER BY id)
FROM {extends} WHERE appid > 0;

INSERT INTO {app_record_ids} (appid, entity, record_id, local_id)
SELECT appid, 'extend_files', id, ROW_NUMBER() OVER (PARTITION BY appid ORDER BY id)
FROM {extend_files} WHERE appid > 0;

INSERT INTO {app_record_ids} (appid, entity, record_id, local_id)
SELECT appid, 'blocks', id, ROW_NUMBER() OVER (PARTITION BY appid ORDER BY id)
FROM {blocks} WHERE appid > 0;

INSERT INTO {app_record_ids} (appid, entity, record_id, local_id)
SELECT appid, 'encrypts', id, ROW_NUMBER() OVER (PARTITION BY appid ORDER BY id)
FROM {encrypts} WHERE appid > 0;

INSERT INTO {app_record_ids} (appid, entity, record_id, local_id)
SELECT appid, 'functions', id, ROW_NUMBER() OVER (PARTITION BY appid ORDER BY id)
FROM {functions} WHERE appid > 0;

INSERT INTO {app_record_counters} (appid, entity, last_id)
SELECT appid, entity, MAX(local_id) FROM {app_record_ids} GROUP BY appid, entity;
