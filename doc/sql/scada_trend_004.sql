
--========================================================================
-- ТРЕНДЫ
--========================================================================

--========================================================================
-- 1. Базовая структура БД
--========================================================================

-- Удаляем все существующие объекты в правильном порядке (зависимости)
DROP TABLE IF EXISTS tags_data CASCADE;
DROP TABLE IF EXISTS tags_info CASCADE;

-- Таблица метаданных тэгов
CREATE TABLE tags_info (
    id BIGSERIAL PRIMARY KEY,
    enable BOOLEAN DEFAULT true,
    tag VARCHAR(500) NOT NULL UNIQUE,
    name VARCHAR(500),
    folder VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Основная таблица с партиционированием
CREATE TABLE tags_data (
    id BIGSERIAL,
    id_obj INTEGER NOT NULL,
    value DOUBLE PRECISION,
    quality INTEGER,
    dt BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (id, dt)
) PARTITION BY RANGE (dt);

--========================================================================
-- 2. Автоматическое управление партициями
--========================================================================

-- Функция создания партиции на день
CREATE OR REPLACE FUNCTION create_partition_for_day(day_date DATE DEFAULT NULL)
RETURNS TEXT AS $$
DECLARE
    target_date DATE := COALESCE(day_date, CURRENT_DATE);
    partition_name TEXT;
    start_dt BIGINT;
    end_dt BIGINT;
BEGIN
    partition_name := 'tags_data_' || TO_CHAR(target_date, 'YYYY_MM_DD');
    start_dt := (EXTRACT(EPOCH FROM target_date) * 1000)::BIGINT;
    end_dt := start_dt + (24 * 60 * 60 * 1000);
    
    IF NOT EXISTS (SELECT 1 FROM pg_tables WHERE tablename = partition_name) THEN
        EXECUTE format(
            'CREATE TABLE %I PARTITION OF tags_data FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_dt, end_dt
        );
        
        -- Создаем индексы для каждой партиции
        EXECUTE format('CREATE INDEX ON %I (dt)', partition_name);
        EXECUTE format('CREATE UNIQUE INDEX ON %I (id_obj, dt)', partition_name);
        
        RAISE NOTICE 'Created partition: %', partition_name;
    END IF;
    
    RETURN partition_name;
END;
$$ LANGUAGE plpgsql;

-- Функция создания партиций на период
CREATE OR REPLACE FUNCTION create_partitions_for_period(start_date DATE, end_date DATE)
RETURNS INTEGER AS $$
DECLARE
    current_dt DATE := start_date;
    created_count INTEGER := 0;
BEGIN
    WHILE current_dt <= end_date LOOP
        PERFORM create_partition_for_day(current_dt);
        created_count := created_count + 1;
        current_dt := current_dt + 1;
    END LOOP;
    
    RETURN created_count;
END;
$$ LANGUAGE plpgsql;

-- Создаем начальные партиции
SELECT create_partitions_for_period(CURRENT_DATE, CURRENT_DATE + 7);

--========================================================================
-- 2.1. Индексы для родительской таблицы (только не-UNIQUE)
--========================================================================

-- Только обычные индексы на родительской таблице
--CREATE INDEX idx_tags_data_dt ON tags_data(dt);
--CREATE INDEX idx_tags_data_id_obj_dt ON tags_data(id_obj, dt);
-- Оставляем обычный индекс на dt для фильтрации по времени
DROP INDEX IF EXISTS idx_tags_data_dt;
CREATE INDEX idx_tags_data_dt ON tags_data(dt);

-- Удаляем обычные индексы и создаем уникальный
DROP INDEX IF EXISTS idx_tags_data_id_obj_dt;
CREATE UNIQUE INDEX idx_tags_data_id_obj_dt ON tags_data(id_obj, dt);


--========================================================================
-- 3. Пакетная вставка данных через временную таблицу
--========================================================================
-- CREATE OR REPLACE FUNCTION sinkross_insert_mess_batch(data_json JSONB)
-- RETURNS JSONB AS $$
-- DECLARE
--     inserted_count INTEGER := 0;
--     total_count INTEGER := 0;
--     start_dt TIMESTAMPTZ := clock_timestamp();
-- BEGIN
--     -- Создаем сегодняшнюю партицию
--     PERFORM create_partition_for_day();
--     
--     -- Получаем общее количество записей
--     SELECT COUNT(*) INTO total_count FROM jsonb_array_elements(data_json);
--     
--     -- Создаем временную таблицу без индексов
--     CREATE TEMP TABLE temp_tags_data ON COMMIT DROP AS
--     SELECT 
--         (elem->>'id_obj')::INTEGER as id_obj,
--         (elem->>'value')::DOUBLE PRECISION as value,
--         (elem->>'quality')::INTEGER as quality,
--         (elem->>'dt')::BIGINT as dt
--     FROM jsonb_array_elements(data_json) AS elem;
--     
--     -- Вставляем только уникальные записи которых нет в основной таблице
--    WITH unique_data AS (
--         SELECT t.* 
--         FROM temp_tags_data t
--         WHERE NOT EXISTS (
--             SELECT 1 FROM tags_data td 
--             WHERE td.id_obj = t.id_obj AND td.dt = t.dt
--         )
--     ),
--     inserted AS (
--         INSERT INTO tags_data (id_obj, value, quality, dt)
--         SELECT id_obj, value, quality, dt
--         FROM unique_data
--         RETURNING 1
--     )
--     SELECT COUNT(*) INTO inserted_count FROM inserted;
--     
--     -- Логируем информацию только при наличии конфликтов
--     IF inserted_count < total_count THEN
--         RAISE LOG 'Inserted % out of % trend records, % duplicates ignored', 
--             inserted_count, total_count, (total_count - inserted_count);
--     END IF;
--     
--     RETURN jsonb_build_object(
--         'inserted', inserted_count,
--         'total', total_count,
--         'duplicates_ignored', (total_count - inserted_count),
--         'execution_dt_ms', EXTRACT(EPOCH FROM (clock_timestamp() - start_dt)) * 1000
--     );
-- END;
-- $$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sinkross_insert_mess_batch(data_json JSONB)
RETURNS JSONB AS $$
DECLARE
    inserted_count INTEGER := 0;
    batch_size INTEGER;
    start_dt TIMESTAMPTZ := clock_timestamp();
BEGIN
    -- Создаем сегодняшнюю партицию
    PERFORM create_partition_for_day();
    
    -- Вставляем данные пачками по 1000 записей
    WITH inserted AS (
        INSERT INTO tags_data (id_obj, value, quality, dt)
        SELECT 
            (elem->>'id_obj')::INTEGER,
            (elem->>'value')::DOUBLE PRECISION,
            (elem->>'quality')::INTEGER,
            (elem->>'dt')::BIGINT
        FROM jsonb_array_elements(data_json) AS elem
		ON CONFLICT (id_obj, dt) DO NOTHING  -- Игнорируем дубликаты по уникальному индексу
        RETURNING 1
    )
    SELECT COUNT(*) INTO inserted_count FROM inserted;
    
    RETURN jsonb_build_object(
        'inserted', inserted_count,
        'execution_dt_ms', EXTRACT(EPOCH FROM (clock_timestamp() - start_dt)) * 1000
    );
END;
$$ LANGUAGE plpgsql;

--========================================================================
-- 4. Усовершенствованная выборка с агрегацией
--========================================================================

CREATE OR REPLACE FUNCTION sinkross_histmess_getdata_json(params_json JSONB)
RETURNS TABLE(
    id BIGINT,
    id_obj INTEGER, 
    value DOUBLE PRECISION,
    quality INTEGER,
    dt BIGINT
) AS $$
DECLARE
    target_id_obj INTEGER;
    dt_start BIGINT;
    dt_end BIGINT;
    agg_type INTEGER;
    limit_count INTEGER;
    max_period_days BIGINT := 30; -- По умолчанию 1 месяц
BEGIN
    target_id_obj := (params_json->>'id_obj')::INTEGER;
    dt_start := (params_json->>'dt_start')::BIGINT;
    dt_end := (params_json->>'dt_end')::BIGINT;
    agg_type := COALESCE((params_json->>'type')::INTEGER, 0);
    limit_count := COALESCE((params_json->>'limit')::INTEGER, 1000);
    max_period_days := COALESCE((params_json->>'max_period_days')::BIGINT, 30);

    --RAISE NOTICE 'target_id_obj: %, dt_start: %, dt_end: %, agg_type: %, limit_count: %', target_id_obj, dt_start,dt_end ,agg_type ,limit_count;

    -- Проверка периода
 	IF (max_period_days) > (365) THEN
        RAISE EXCEPTION 'Period % too long. Maximum 365 days allowed.', max_period_days;
    END IF;

    IF (dt_end - dt_start) > (max_period_days * 24 * 60 * 60 * 1000) THEN
        RAISE EXCEPTION 'Period too long. Maximum % days allowed.', max_period_days;
    END IF;

    
    RETURN QUERY 
    SELECT * FROM get_aggregated_data(target_id_obj, dt_start, dt_end, agg_type, limit_count);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_aggregated_data(
    p_id_obj INTEGER,
    p_dt_start BIGINT,
    p_dt_end BIGINT,
    p_agg_type INTEGER,
    p_limit INTEGER
)
RETURNS TABLE(
    id BIGINT,
    id_obj INTEGER,
    value DOUBLE PRECISION, 
    quality INTEGER,
    dt BIGINT
) AS $$
DECLARE
    total_records BIGINT;
    step_size INTEGER; -- Количество записей в одном шаге
    bucket_size INTEGER; -- Размер бакета для мин-макс
BEGIN
    -- Получаем общее количество записей
    EXECUTE '
        SELECT COUNT(*) 
        FROM tags_data 
        WHERE id_obj = $1 AND dt BETWEEN $2 AND $3'
    INTO total_records
    USING p_id_obj, p_dt_start, p_dt_end;

    --RAISE NOTICE 'total_records: %, p_id_obj: %', total_records, p_id_obj;
    
    IF total_records <= p_limit THEN
        -- Данные помещаются в лимит - возвращаем как есть
        RETURN QUERY EXECUTE '
            SELECT id, id_obj, value, quality, dt
            FROM tags_data 
            WHERE id_obj = $1 AND dt BETWEEN $2 AND $3
            ORDER BY dt
            LIMIT $4'
        USING p_id_obj, p_dt_start, p_dt_end, p_limit;
    ELSE
        -- Вычисляем шаг для разных алгоритмов
        CASE p_agg_type
		    WHEN 1 THEN -- Среднее: берем каждую N-ю запись
		        step_size := GREATEST(1, (total_records / p_limit)::INTEGER);
		        --RAISE NOTICE 'p_id_obj: %, p_dt_start: %, p_dt_end: %, step_size: %, p_limit: %', p_id_obj, p_dt_start, p_dt_end, step_size, p_limit;
		        
		        RETURN QUERY EXECUTE '
		            WITH numbered_records AS (
		                SELECT id, id_obj, value, quality, dt,
		                       ROW_NUMBER() OVER (ORDER BY dt) as rn
		                FROM tags_data 
		                WHERE id_obj = $1 AND dt BETWEEN $2 AND $3
		            )
		            SELECT id, id_obj, value, quality, dt
		            FROM numbered_records
		            WHERE (rn - 1) % $4 = 0  -- Берем каждую step_size-ю запись, начиная с первой
		            ORDER BY dt
		            LIMIT $5'
		        USING p_id_obj, p_dt_start, p_dt_end, step_size, p_limit;
		                
            WHEN 2 THEN -- Мин-Макс: в каждом интервале берем мин и макс
                bucket_size := (total_records / (p_limit / 2))::INTEGER;
                
                RETURN QUERY EXECUTE '
                    WITH numbered_records AS (
                        SELECT id, id_obj, value, quality, dt,
                               ROW_NUMBER() OVER (ORDER BY dt) as rn,
                               ((ROW_NUMBER() OVER (ORDER BY dt) - 1) / $4)::INTEGER as bucket
                        FROM tags_data 
                        WHERE id_obj = $1 AND dt BETWEEN $2 AND $3
                    ),
                    bucket_extremes AS (
                        SELECT 
                            bucket,
                            MIN(value) as min_value,
                            MAX(value) as max_value,
                            MIN(quality) as min_quality,
                            MIN(dt) as min_dt,
                            MAX(dt) as max_dt
                        FROM numbered_records
                        GROUP BY bucket
                    )
                    SELECT 
                        ROW_NUMBER() OVER (ORDER BY bucket, value_type) as id,
                        $1 as id_obj,
                        value,
                        min_quality as quality,
                        dt
                    FROM (
                        -- Минимальные значения
                        SELECT 
                            bucket, min_value as value, min_quality,
                            min_dt as dt, ''min'' as value_type
                        FROM bucket_extremes
                        UNION ALL
                        -- Максимальные значения  
                        SELECT 
                            bucket, max_value as value, min_quality,
                            max_dt as dt, ''max'' as value_type
                        FROM bucket_extremes
                    ) extremes
                    ORDER BY dt
                    LIMIT $5'
                USING p_id_obj, p_dt_start, p_dt_end, bucket_size, p_limit;
                
            WHEN 3 THEN -- Минимум: в каждом интервале берем только минимум
                step_size := (total_records / p_limit)::INTEGER;
                
                RETURN QUERY EXECUTE '
                    WITH numbered_records AS (
                        SELECT id, id_obj, value, quality, dt,
                               ROW_NUMBER() OVER (ORDER BY dt) as rn,
                               ((ROW_NUMBER() OVER (ORDER BY dt) - 1) / $4)::INTEGER as bucket
                        FROM tags_data 
                        WHERE id_obj = $1 AND dt BETWEEN $2 AND $3
                    )
                    SELECT 
                        MIN(id) as id,
                        $1 as id_obj,
                        MIN(value) as value,
                        MIN(quality) as quality,
                        MIN(dt) as dt
                    FROM numbered_records
                    GROUP BY bucket
                    ORDER BY dt
                    LIMIT $5'
                USING p_id_obj, p_dt_start, p_dt_end, step_size, p_limit;
                
            WHEN 4 THEN -- Максимум: в каждом интервале берем только максимум
                step_size := (total_records / p_limit)::INTEGER;
                
                RETURN QUERY EXECUTE '
                    WITH numbered_records AS (
                        SELECT id, id_obj, value, quality, dt,
                               ROW_NUMBER() OVER (ORDER BY dt) as rn,
                               ((ROW_NUMBER() OVER (ORDER BY dt) - 1) / $4)::INTEGER as bucket
                        FROM tags_data 
                        WHERE id_obj = $1 AND dt BETWEEN $2 AND $3
                    )
                    SELECT 
                        MIN(id) as id,
                        $1 as id_obj,
                        MAX(value) as value,
                        MIN(quality) as quality,
                        MIN(dt) as dt
                    FROM numbered_records
                    GROUP BY bucket
                    ORDER BY dt
                    LIMIT $5'
                USING p_id_obj, p_dt_start, p_dt_end, step_size, p_limit;
                
            ELSE -- Все записи (первые p_limit)
                RETURN QUERY EXECUTE '
                    SELECT id, id_obj, value, quality, dt
                    FROM tags_data 
                    WHERE id_obj = $1 AND dt BETWEEN $2 AND $3
                    ORDER BY dt
                    LIMIT $4'
                USING p_id_obj, p_dt_start, p_dt_end, p_limit;
        END CASE;
    END IF;
END;
$$ LANGUAGE plpgsql;


--========================================================================
-- 5. Управление данными и очистка
--========================================================================

-- Функция удаления старых данных
CREATE OR REPLACE FUNCTION cleanup_old_data(retention_months INTEGER DEFAULT 12)
RETURNS JSONB AS $$
DECLARE
    cutoff_date DATE := CURRENT_DATE - (retention_months * INTERVAL '1 month');
    cutoff_timestamp BIGINT;
    deleted_tables TEXT[];
    table_name TEXT;
BEGIN
    cutoff_timestamp := (EXTRACT(EPOCH FROM cutoff_date) * 1000)::BIGINT;
    
    -- Находим и удаляем старые партиции
    FOR table_name IN (
        SELECT inhrelid::regclass::text 
        FROM pg_inherits 
        WHERE inhparent = 'tags_data'::regclass
        AND (regexp_match(inhrelid::regclass::text, 'tags_data_(\d{4})_(\d{2})_(\d{2})'))[1]::integer < EXTRACT(YEAR FROM cutoff_date)
        OR (
            (regexp_match(inhrelid::regclass::text, 'tags_data_(\d{4})_(\d{2})_(\d{2})'))[1]::integer = EXTRACT(YEAR FROM cutoff_date)
            AND (regexp_match(inhrelid::regclass::text, 'tags_data_(\d{4})_(\d{2})_(\d{2})'))[2]::integer < EXTRACT(MONTH FROM cutoff_date)
        )
    )
    LOOP
        EXECUTE format('DROP TABLE %I', table_name);
        deleted_tables := array_append(deleted_tables, table_name);
    END LOOP;
    
    RETURN jsonb_build_object(
        'deleted_tables', deleted_tables,
        'retention_months', retention_months,
        'cutoff_date', cutoff_date
    );
END;
$$ LANGUAGE plpgsql;

-- Функция сброса к default настройкам
CREATE OR REPLACE FUNCTION reset_to_default()
RETURNS JSONB AS $$
DECLARE
    table_name TEXT;
BEGIN
    -- Удаляем все партиции кроме текущей и следующих 7 дней
    FOR table_name IN (
        SELECT inhrelid::regclass::text 
        FROM pg_inherits 
        WHERE inhparent = 'tags_data'::regclass
        AND inhrelid::regclass::text != 'tags_data_' || TO_CHAR(CURRENT_DATE, 'YYYY_MM_DD')
        AND (regexp_match(inhrelid::regclass::text, 'tags_data_(\d{4})_(\d{2})_(\d{2})'))[1]::text || '_' || 
            (regexp_match(inhrelid::regclass::text, 'tags_data_(\d{4})_(\d{2})_(\d{2})'))[2]::text || '_' ||
            (regexp_match(inhrelid::regclass::text, 'tags_data_(\d{4})_(\d{2})_(\d{2})'))[3]::text < 
            TO_CHAR(CURRENT_DATE - INTERVAL '7 days', 'YYYY_MM_DD')
    )
    LOOP
        EXECUTE format('DROP TABLE %I', table_name);
    END LOOP;
    
    -- Создаем партиции на ближайшие 7 дней
    PERFORM create_partitions_for_period(CURRENT_DATE, CURRENT_DATE + 7);
    
    RETURN jsonb_build_object('status', 'reset_complete');
END;
$$ LANGUAGE plpgsql;

--========================================================================
-- 6. Тестовые данные и функции
--========================================================================

-- Функция генерации тестовых данных с пакетной вставкой
CREATE OR REPLACE FUNCTION generate_test_data(
    obj_count INTEGER DEFAULT 10,
    hours_back INTEGER DEFAULT 1, 
    points_per_hour INTEGER DEFAULT 60
)
RETURNS JSONB AS $$
DECLARE
    result JSONB;
    start_dt BIGINT;
    end_dt BIGINT;
    dt_step BIGINT;
    current_dt BIGINT;
    i INTEGER;
    j INTEGER;
    total_points INTEGER := 0;
    batch_points JSONB;
    batch_size INTEGER := 1000; -- Вставляем батчами по 1000 записей
    current_batch JSONB := '[]'::JSONB;
    batch_count INTEGER := 0;
BEGIN
    start_dt := (EXTRACT(EPOCH FROM NOW() - (hours_back * INTERVAL '1 hour')) * 1000)::BIGINT;
    end_dt := (EXTRACT(EPOCH FROM NOW()) * 1000)::BIGINT;
    dt_step := (60 * 60 * 1000) / points_per_hour;
    
    RAISE NOTICE 'Generating test data: % objects, % hours back, % points per hour', 
                 obj_count, hours_back, points_per_hour;
    RAISE NOTICE 'Time range: % to %', start_dt, end_dt;
    
    FOR i IN 1..obj_count LOOP
        RAISE NOTICE 'Generating data for object %', i;
        current_dt := start_dt;
        
        WHILE current_dt <= end_dt LOOP
            -- Добавляем точку в текущий батч
            current_batch := current_batch || jsonb_build_object(
                'id_obj', i,
                'value', (random() * 100)::numeric(10,4),
                'quality', CASE WHEN random() > 0.05 THEN 1 ELSE 0 END,
                'dt', current_dt
            );
            
            total_points := total_points + 1;
            
            -- Если батч заполнен, вставляем его
            IF jsonb_array_length(current_batch) >= batch_size THEN
                PERFORM sinkross_insert_mess_batch(current_batch);
                batch_count := batch_count + 1;
                current_batch := '[]'::JSONB;
                RAISE NOTICE 'Inserted batch %, total points: %', batch_count, total_points;
            END IF;
            
            current_dt := current_dt + dt_step;
            -- RAISE NOTICE 'Inserted batch %, total points: %', current_dt, dt_step;
        END LOOP;
    END LOOP;
    
    -- Вставляем оставшиеся данные
    IF jsonb_array_length(current_batch) > 0 THEN
        PERFORM sinkross_insert_mess_batch(current_batch);
        batch_count := batch_count + 1;
        RAISE NOTICE 'Inserted final batch %, total points: %', batch_count, total_points;
    END IF;
    
    RETURN jsonb_build_object(
        'generated_points', total_points,
        'batches_inserted', batch_count,
        'objects_count', obj_count,
        'time_range_hours', hours_back,
        'points_per_hour', points_per_hour
    );
END;
$$ LANGUAGE plpgsql;

--========================================================================
-- 7. Дополнительное обслуживание:
--========================================================================

-- Переиндексация (выполняется отдельно для каждого индекса)
CREATE OR REPLACE FUNCTION reindex_trend_tables()
RETURNS JSONB AS $$
DECLARE
    start_time TIMESTAMPTZ := clock_timestamp();
BEGIN
    RAISE NOTICE 'Starting reindex...';
    
    -- Исправленные имена индексов (те, которые реально существуют)
    REINDEX INDEX idx_tags_data_dt;
    RAISE NOTICE 'Index idx_tags_data_dt reindexed';
    
    REINDEX INDEX idx_tags_data_id_obj_dt;  -- ИСПРАВЛЕНО: правильное имя
    RAISE NOTICE 'Index idx_tags_data_id_obj_dt reindexed';
    
    RETURN jsonb_build_object(
        'status', 'success',
        'execution_time_ms', EXTRACT(EPOCH FROM (clock_timestamp() - start_time)) * 1000,
        'message', 'Reindex completed successfully'
    );
EXCEPTION
    WHEN others THEN
        RETURN jsonb_build_object(
            'status', 'error',
            'error_message', SQLERRM,
            'execution_time_ms', EXTRACT(EPOCH FROM (clock_timestamp() - start_time)) * 1000
        );
END;
$$ LANGUAGE plpgsql;

-- Автоматическое обслуживание (запускать раз в день)
CREATE OR REPLACE FUNCTION daily_maintenance()
RETURNS JSONB AS $$
DECLARE
    result JSONB;
BEGIN
    -- Создаем партиции на ближайшие 3 дня
    PERFORM create_partitions_for_period(CURRENT_DATE, CURRENT_DATE + 3);
    
    -- Удаляем данные старше 12 месяцев
    SELECT cleanup_old_data(12) INTO result;
    
    -- Анализируем таблицы для оптимизации запросов
    ANALYZE tags_data;
    ANALYZE tags_info;
    
    RETURN jsonb_build_object(
        'maintenance_complete', true,
        'cleanup_result', result
    );
END;
$$ LANGUAGE plpgsql;

-- Статистика по использованию
CREATE OR REPLACE FUNCTION get_storage_stats()
RETURNS TABLE(metric TEXT, value NUMERIC) AS $$
BEGIN
    RETURN QUERY SELECT 'total_partitions', COUNT(*)::NUMERIC 
    FROM pg_inherits WHERE inhparent = 'tags_data'::regclass;
    
    RETURN QUERY SELECT 'total_records', COUNT(*)::NUMERIC FROM tags_data;
    
    RETURN QUERY SELECT 'storage_size_gb', 
        pg_total_relation_size('tags_data') / (1024^3)::NUMERIC;
    
    RETURN QUERY SELECT 'oldest_data_days', 
        EXTRACT(DAY FROM NOW() - MIN(TO_TIMESTAMP(dt/1000)))::NUMERIC
    FROM tags_data;
END;
$$ LANGUAGE plpgsql;

-- Проверка работоспособности:
-- Создать начальные партиции
-- SELECT create_partitions_for_period(CURRENT_DATE, CURRENT_DATE + 3);
-- Проверить структуру
-- SELECT * FROM get_storage_stats();
-- Протестировать вставку
-- select generate_test_data();
-- SELECT * FROM get_aggregated_data(1, 1763287245169, 1763373645169, 1, 10)
-- SELECT * FROM sinkross_histmess_getdata_json(jsonb_build_object( 'id_obj', 1, 'dt_start', 1763287245169, 'dt_end', 1763373645169,  'type', 4,  'limit', 20 ));

--========================================================================
-- 8. Функции для работы с tags_info в JSON формате
--========================================================================

-- Функция для загрузки трендов с сохранением оригинальных ID
CREATE OR REPLACE FUNCTION load_trends_from_json_with_ids(file_path TEXT)
RETURNS JSONB AS $$
DECLARE
    file_content TEXT;
    json_data JSONB;
    inserted_count INTEGER := 0;
    error_count INTEGER := 0;
    total_count INTEGER := 0;
    result JSONB;
BEGIN
    -- Читаем файл
    BEGIN
        SELECT pg_read_file(file_path) INTO file_content;
    EXCEPTION
        WHEN OTHERS THEN
            RETURN jsonb_build_object(
                'status', 'error',
                'message', 'Failed to read file: ' || SQLERRM
            );
    END;
    
    -- Парсим JSON
    BEGIN
        json_data := file_content::JSONB;
    EXCEPTION
        WHEN OTHERS THEN
            RETURN jsonb_build_object(
                'status', 'error',
                'message', 'Invalid JSON format: ' || SQLERRM
            );
    END;
    
    -- Очищаем таблицу
    TRUNCATE TABLE tags_info RESTART IDENTITY CASCADE;
    
    -- Вставляем данные с оригинальными ID
    WITH inserted AS (
        INSERT INTO tags_info (id, enable, tag, name, folder, created_at)
        SELECT 
            (value->>'id')::BIGINT,
            (value->>'enable')::BOOLEAN,
            value->>'tag',
            value->>'name',
            value->>'folder',
            NOW()
        FROM jsonb_each(json_data->'tags')
        ON CONFLICT (id) DO UPDATE SET
            enable = EXCLUDED.enable,
            tag = EXCLUDED.tag,
            name = EXCLUDED.name,
            folder = EXCLUDED.folder,
            created_at = NOW()
        RETURNING 1
    )
    SELECT COUNT(*) INTO inserted_count FROM inserted;
    
    -- Получаем общее количество записей в JSON
    SELECT COUNT(*) INTO total_count 
    FROM jsonb_each(json_data->'tags');
    
    -- Вычисляем количество ошибок
    error_count := total_count - inserted_count;
    
    -- Формируем результат
    result := jsonb_build_object(
        'status', 'success',
        'inserted', inserted_count,
        'total', total_count,
        'errors', error_count,
        'message', 'Loaded ' || inserted_count || ' out of ' || total_count || ' tags'
    );
    
    RETURN result;
    
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'status', 'error',
            'message', 'Unexpected error: ' || SQLERRM
        );
END;
$$ LANGUAGE plpgsql;

-- SELECT * FROM load_trends_from_json_with_ids('E:\!!!VMWARE\VM_GO\windows\project\go-server\server-system\cmd\configs\objects\trend.json');

-- Функция для очистки и заполнения таблицы tags_info данными в JSON формате с оригинальными ID
CREATE OR REPLACE FUNCTION tags_info_load_from_json(data_json JSONB)
RETURNS JSONB AS $$
DECLARE
    inserted_count INTEGER := 0;
    deleted_count INTEGER := 0;
    total_count INTEGER := 0;
BEGIN
    -- Получаем общее количество записей в JSON
    SELECT COUNT(*) INTO total_count 
    FROM jsonb_each(data_json->'tags');
    
    -- Очищаем таблицу
    DELETE FROM tags_info;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Вставляем новые данные с оригинальными ID
    INSERT INTO tags_info (id, enable, tag, name, folder)
    SELECT 
        (value->>'id')::BIGINT,
        COALESCE((value->>'enable')::BOOLEAN, true),
        value->>'tag',
        value->>'name',
        value->>'folder'
    FROM jsonb_each(data_json->'tags') AS t(key, value);
    
    GET DIAGNOSTICS inserted_count = ROW_COUNT;
    
    RETURN jsonb_build_object(
        'deleted_records', deleted_count,
        'inserted_records', inserted_count,
        'total_json_records', total_count,
        'status', 'success'
    );
    
EXCEPTION
    WHEN others THEN
        RETURN jsonb_build_object(
            'status', 'error',
            'error_message', SQLERRM,
            'deleted_records', deleted_count,
            'inserted_records', inserted_count,
            'total_json_records', total_count
        );
END;
$$ LANGUAGE plpgsql;

-- Функция для получения всех данных из tags_info в JSON формате
CREATE OR REPLACE FUNCTION tags_info_get_all_json()
RETURNS JSONB AS $$
DECLARE
    result JSONB;
BEGIN
    SELECT jsonb_agg(jsonb_build_object(
        'id', id,
        'enable', enable,
        'tag', tag,
        'name', name,
        'folder', folder,
        'created_at', created_at
    )) INTO result
    FROM tags_info
    ORDER BY id;
    
    RETURN COALESCE(result, '[]'::JSONB);
END;
$$ LANGUAGE plpgsql;

--========================================================================
-- 9. Критические настройки PostgreSQL для больших объемов:
--========================================================================

-- В postgresql.conf:
-- shared_buffers = 25% от RAM
-- work_mem = 256MB - 1GB
-- maintenance_work_mem = 1GB - 4GB  
-- effective_cache_size = 75% от RAM
-- max_connections = достаточное для вашего приложения
-- checkpoint_timeout = 30min
-- max_wal_size = 10GB - 50GB

--========================================================================
-- 10. Объемы данных для разных масштабов:
--========================================================================

-- Сценарий 1: 1,000 объектов
-- Частота записи: каждые 10 секунд
-- Записей в день: 1,000 × 8,640 = 8.64 млн
-- Размер записи: ~40 байт
-- Ежедневный объем: 8.64M × 40 байт = 345 МБ
-- Годовой объем: 345 МБ × 365 = 126 ГБ
-- Память для выборки: ~100-500 МБ

-- Сценарий 2: 10,000 объектов
-- Записей в день: 10,000 × 8,640 = 86.4 млн
-- Ежедневный объем: 86.4M × 40 байт = 3.46 ГБ
-- Годовой объем: 3.46 ГБ × 365 = 1.26 ТБ
-- Память для выборки: ~1-2 ГБ

-- Сценарий 3: 100,000 объектов
-- Записей в день: 100,000 × 8,640 = 864 млн
-- Ежедневный объем: 864M × 40 байт = 34.6 ГБ
-- Годовой объем: 34.6 ГБ × 365 = 12.6 ТБ
-- Память для выборки: ~4-8 ГБ

-- Время выборки (приблизительно):
-- Тип 0 (все записи): 100-500 мс (зависит от количества данных)
-- Тип 1 (среднее): 200-800 мс
-- Тип 2 (мин-макс): 300-1000 мс

-- За месяц: 1-3 секунды
-- За год: 5-15 секунд

-- Время выборки при увеличении периода:
-- 1 месяц → 3 месяца: время ×2-3
-- 1 месяц → 6 месяцев: время ×4-6
-- 1 месяц → 12 месяцев: время ×8-12

