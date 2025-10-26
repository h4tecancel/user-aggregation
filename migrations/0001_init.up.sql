CREATE TABLE IF NOT EXISTS user_info (
    user_id      uuid        NOT NULL,
    service_name text        NOT NULL,
    price        bigint      NOT NULL,
    start_date   timestamptz NOT NULL,
    end_date     timestamptz NOT NULL,
    PRIMARY KEY (user_id, service_name, start_date)
);

CREATE INDEX IF NOT EXISTS idx_user_info_user_id
  ON user_info (user_id);

CREATE INDEX IF NOT EXISTS idx_user_info_user_service
  ON user_info (user_id, service_name);

-- Опционально: индекс по датам, если часто фильтруешь по интервалам
CREATE INDEX IF NOT EXISTS idx_user_info_start_end
  ON user_info (start_date, end_date);