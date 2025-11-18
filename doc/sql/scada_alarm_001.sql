
--========================================================================
-- 1. ИСТОРИЧЕСКИЙ ЖУРНАЛ (sinkross_histmess)
--========================================================================

-- 1. Удаляем все существующие объекты в правильном порядке (зависимости)
DROP FUNCTION IF EXISTS sinkross_insert_mess_batch CASCADE;
DROP INDEX IF EXISTS idx_sinkross_histmess_dt CASCADE;
DROP INDEX IF EXISTS idx_sinkross_histmess_code_dt CASCADE;
DROP TABLE IF EXISTS sinkross_histmess CASCADE;
DROP TYPE IF EXISTS alarm_mess_type CASCADE;

-- 2. Создаем таблицу для хранения данных
CREATE TABLE IF NOT EXISTS sinkross_histmess (
    id           BIGSERIAL PRIMARY KEY,
    code         BIGINT,
    dt           BIGINT,
    dt_txt       text,
	tag          text,
    mess_full    text,
	mess_name    text,
	mess_state   text,
	uso_id		 int,
	uso_txt		 text,
	users        text,  
	severity     int,
    opermess     int,
    color        text,
	kvit         boolean,
	dt_kvit      BIGINT,
    dt_kvit_txt  text
);

-- 3. Создаем индекс
CREATE INDEX idx_sinkross_histmess_dt ON sinkross_histmess(dt);
CREATE UNIQUE INDEX idx_sinkross_histmess_code_dt ON sinkross_histmess(code, dt);

-- 4. Создаем составной тип, принимающий данные как есть
CREATE TYPE alarm_mess_type AS (
    code         BIGINT,
    dt           BIGINT,
    dt_txt       text,
	tag          text,
    mess_full    text,
	mess_name    text,
	mess_state   text,
	uso_id		 int,
	uso_txt		 text,
	users        text,  
	severity     int,
    opermess     int,
    color        text
);

-- 5. Функция для пакетной вставки из JSON
CREATE OR REPLACE FUNCTION sinkross_insert_mess_batch(
    p_messages JSONB
)
RETURNS INTEGER AS $$
DECLARE
    v_count INTEGER;
BEGIN
    -- Вставка из JSON
    WITH inserted AS (
        INSERT INTO sinkross_histmess (
            code, dt, dt_txt, tag, mess_full, mess_name, mess_state, 
            uso_id, uso_txt, users, severity, opermess, color, kvit
        )
        SELECT 
            (msg->>'code')::BIGINT,
            (msg->>'dt')::BIGINT,                        
            msg->>'dt_txt',
            msg->>'tag',
            msg->>'mess_full',                   
            msg->>'mess_name',                 
            msg->>'mess_state',                 
            (msg->>'uso_id')::INTEGER, 
            msg->>'uso_txt',       			
            msg->>'users',
            (msg->>'severity')::INTEGER,                  
            (msg->>'opermess')::INTEGER,                      
            msg->>'color',                     
            false
        FROM jsonb_array_elements(p_messages) AS msg
        ON CONFLICT (code, dt) DO NOTHING
        RETURNING 1
    )
    SELECT COUNT(*) INTO v_count FROM inserted;
    
    RETURN v_count;
EXCEPTION
    WHEN OTHERS THEN
        INSERT INTO sinkross_errorlog(message)
        VALUES('Batch insert error: ' || SQLERRM);
        RETURN -1;
END;
$$ LANGUAGE plpgsql;


--========================================================================
-- 2. ТАБЛИЦА ДЛЯ СООБЩЕНИЙ ОБ ОШИБКАХ РАБОТЫ С БД (sinkross_errorlog)
--========================================================================

-- 1. Удаляем все существующие объекты в правильном порядке (зависимости)
DROP TABLE IF EXISTS sinkross_errorlog;

-- 2. Создаем таблицу для хранения данных
CREATE TABLE IF NOT EXISTS sinkross_errorlog (
	id			BIGSERIAL PRIMARY KEY,
	dttime		bigint not null default ((extract(epoch from now()) * 10000000) + 116444736000000000),
	message		text NOT NULL
);


--========================================================================
-- 3. Функция для получения N записей
--========================================================================

-- 2. Функция для получения N записей
CREATE OR REPLACE FUNCTION sinkross_histmess_getdata_json(
    params_json JSONB
)
RETURNS TABLE(
    id BIGINT,
    code BIGINT,
    dt BIGINT,
    dt_txt TEXT,
    tag TEXT,
    mess_full TEXT,
    mess_name TEXT,
    mess_state TEXT,
    uso_id INT,
    uso_txt TEXT,
    users TEXT,
    severity INT,
    opermess INT,
    color TEXT,
    kvit BOOLEAN,
    dt_kvit BIGINT,
    dt_kvit_txt TEXT,
    current_page INT,
    total_pages INT
) AS $$
DECLARE
    dt_start BIGINT;
    dt_end BIGINT;
    tag_find TEXT;
    mess_full_find TEXT;
    uso_txt_find TEXT;
    severity_find INT;
    opermess_find INT;
    kvit_find INT;
    page_num INT;
    
    query_text TEXT;
    where_clause TEXT := '';
    count_query TEXT;
    total_records BIGINT;
    max_page INT;
BEGIN
    -- Извлекаем параметры из JSON
    dt_start := COALESCE((params_json->>'dt_start')::BIGINT, NULL);
    dt_end := COALESCE((params_json->>'dt_end')::BIGINT, NULL);
    tag_find := COALESCE(params_json->>'tag_find', '');
    mess_full_find := COALESCE(params_json->>'mess_full_find', '');
    uso_txt_find := COALESCE(params_json->>'uso_txt_find', '');
    severity_find := COALESCE((params_json->>'severity_find')::INT, 0);
    opermess_find := COALESCE((params_json->>'opermess_find')::INT, 0);
    kvit_find := COALESCE((params_json->>'kvit_find')::INT, 0);
    page_num := COALESCE((params_json->>'page_num')::INT, 1);

    -- Формируем условия WHERE
    IF dt_start IS NOT NULL AND dt_start > 0 THEN
        where_clause := where_clause || ' AND dt >= ' || dt_start;
    END IF;
    
    IF dt_end IS NOT NULL AND dt_end > 0 THEN
        where_clause := where_clause || ' AND dt <= ' || dt_end;
    END IF;
    
    IF tag_find != '' THEN
        where_clause := where_clause || ' AND tag ILIKE ''%' || tag_find || '%''';
    END IF;
    
    IF mess_full_find != '' THEN
        where_clause := where_clause || ' AND mess_full ILIKE ''%' || mess_full_find || '%''';
    END IF;
    
    IF uso_txt_find != '' THEN
        where_clause := where_clause || ' AND uso_txt ILIKE ''%' || uso_txt_find || '%''';
    END IF;
    
    IF severity_find > 0 THEN
        where_clause := where_clause || ' AND severity = ' || severity_find;
    END IF;
    
    IF opermess_find = 1 THEN
        where_clause := where_clause || ' AND opermess = 1';
    END IF;
    
    IF kvit_find = 1 THEN
        where_clause := where_clause || ' AND kvit = false';
    ELSIF kvit_find = 2 THEN
        where_clause := where_clause || ' AND kvit = true';
    END IF;
    
    -- Удаляем начальный " AND " если условия есть
    IF length(where_clause) > 0 THEN
        where_clause := ' WHERE ' || substring(where_clause from 6);
    END IF;
    
    -- Получаем общее количество записей
    count_query := 'SELECT COUNT(*) FROM sinkross_histmess ' || where_clause;
    EXECUTE count_query INTO total_records;
    
    -- Вычисляем общее количество страниц
    total_pages := CEIL(total_records::numeric / 100);
    IF total_pages = 0 THEN
        total_pages := 1;
    END IF;
    
    -- Проверяем номер страницы
    IF page_num <= 0 OR page_num > total_pages THEN
        current_page := 1;
    ELSE
        current_page := page_num;
    END IF;
    
    -- Формируем основной запрос
    query_text := 'SELECT 
        id, code, dt, dt_txt, tag, mess_full, mess_name, mess_state, 
        uso_id, uso_txt, users, severity, opermess, color, kvit, 
        dt_kvit, dt_kvit_txt,
        ' || current_page || ' AS current_page,
        ' || total_pages || ' AS total_pages
        FROM sinkross_histmess ' || where_clause || 
        ' ORDER BY id DESC LIMIT 100 OFFSET ' || ((current_page - 1) * 100);

    RAISE NOTICE 'Main query: %', query_text;

    -- Выполняем запрос
    RETURN QUERY EXECUTE query_text;
END;
$$ LANGUAGE plpgsql;


--========================================================================
-- 4. Тестовая функция для генерации данных (JSON версия)
--========================================================================
-- Функция для тестового заполнения данными
CREATE OR REPLACE FUNCTION sinkross_test_fill_batch(
    p_count INTEGER DEFAULT 100
)
RETURNS TABLE(
    inserted_count INTEGER,
    test_details TEXT
) AS $$
DECLARE
    v_test_data JSONB;
    v_inserted INTEGER;
    v_start_time BIGINT;
    v_end_time BIGINT;
    v_current_filetime BIGINT;
    v_base_filetime BIGINT;
    v_i INTEGER;
    v_json_array JSONB := '[]'::JSONB;
BEGIN
    -- Получаем текущее время в формате filetime
    v_current_filetime := (EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 10000000) + 116444736000000000;
    v_base_filetime := v_current_filetime - (p_count * 10000000); -- Отступаем назад для тестовых данных
    
    -- Генерируем тестовые данные в формате JSON
    FOR v_i IN 1..p_count LOOP
        v_json_array := v_json_array || jsonb_build_object(
            'code', (1000 + v_i),
            'dt', (v_base_filetime + (v_i * 10000000)),
            'dt_txt', sinkross_filetime_to_text(v_base_filetime + (v_i * 10000000)),
            'tag', 'TAG_' || (v_i % 10 + 1),
            'mess_full', 'Полное сообщение для теста №' || v_i,
            'mess_name', 'Имя сообщения ' || v_i,
            'mess_state', CASE WHEN v_i % 3 = 0 THEN 'Авария' 
                             WHEN v_i % 3 = 1 THEN 'Предупреждение' 
                             ELSE 'Норма' END,
            'uso_id', (v_i % 5 + 1),
            'uso_txt', 'USO_TXT_' || (v_i % 5 + 1),
            'users', 'user' || (v_i % 3 + 1),
            'severity', (v_i % 4 + 1),
            'opermess', CASE WHEN v_i % 4 = 0 THEN 1 ELSE 0 END,
            'color', CASE WHEN v_i % 4 = 0 THEN '#FF0000' 
                        WHEN v_i % 4 = 1 THEN '#FFFF00' 
                        WHEN v_i % 4 = 2 THEN '#00FF00' 
                        ELSE '#0000FF' END
        );
    END LOOP;

    v_test_data := v_json_array;

    -- Записываем время начала
    v_start_time := (EXTRACT(EPOCH FROM clock_timestamp()) * 10000000) + 116444736000000000;
    
    -- Вызываем пакетную вставку с JSON данными
    SELECT sinkross_insert_mess_batch(v_test_data) INTO v_inserted;
    
    -- Записываем время окончания
    v_end_time := (EXTRACT(EPOCH FROM clock_timestamp()) * 10000000) + 116444736000000000;
    
    -- Возвращаем результаты
    INSERTED_COUNT := v_inserted;
    TEST_DETAILS := 'Вставлено ' || v_inserted || ' из ' || p_count || ' записей. ' ||
                   'Время выполнения: ' || ((v_end_time - v_start_time) / 10000) || ' мс';
    
    RETURN NEXT;
    
EXCEPTION
    WHEN OTHERS THEN
        INSERTED_COUNT := -1;
        TEST_DETAILS := 'Ошибка: ' || SQLERRM;
        RETURN NEXT;
END;
$$ LANGUAGE plpgsql;



-- Функция для очистки тестовых данных
CREATE OR REPLACE FUNCTION sinkross_clean_test_data()
RETURNS TEXT AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    -- Удаляем тестовые данные (по определенным тегам или кодам)
    DELETE FROM sinkross_histmess 
    WHERE tag LIKE 'TAG_%' 
       OR tag = 'DUPLICATE_TAG'
       OR code BETWEEN 1000 AND 3000;
    
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    
    RETURN 'Удалено тестовых записей: ' || v_deleted_count;
    
EXCEPTION
    WHEN OTHERS THEN
        RETURN 'Ошибка при очистке: ' || SQLERRM;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION sinkross_filetime_to_text(filetime BIGINT) 
RETURNS TEXT AS $$
DECLARE
    seconds_part BIGINT;
    fractions_part BIGINT;
    base_time TIMESTAMP;
    formatted_time TEXT;
BEGIN
    IF filetime IS NULL THEN
        RETURN NULL;
    END IF;
    
    seconds_part := filetime / 10000000;
    fractions_part := filetime % 10000000;
    base_time := '1601-01-01'::timestamp + (seconds_part * INTERVAL '1 second');
    
    formatted_time := TO_CHAR(base_time, 'DD.MM.YYYY HH24:MI:SS') || 
                     '.' || LPAD(fractions_part::text, 3, '0');
    
    RETURN formatted_time;
END;
$$ LANGUAGE plpgsql;

-- Тест с 50 записями
-- SELECT * FROM sinkross_histmess_getdata_json('{"page_num": 3}'::jsonb);
-- SELECT * FROM sinkross_histmess_getdata_json('{}'::jsonb);
-- SELECT * FROM sinkross_histmess_getdata_json('{"tag_find": "01"}'::jsonb);
-- SELECT * FROM sinkross_histmess_getdata_json('{"dt_start": 1763106060000,"dt_end": 1763109660000}'::jsonb);
