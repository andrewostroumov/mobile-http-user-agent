CREATE TABLE IF NOT EXISTS devices (
    id serial primary key,
    manufacturer varchar,
    model varchar,
    build varchar,
    cpu_description varchar,
    display_x integer,
    display_y integer,
    base_android_version double precision,
    dpi integer,
    build_os_device varchar
);

ALTER TABLE devices DROP CONSTRAINT IF EXISTS devices_fingerprint_uniq;
CREATE UNIQUE INDEX IF NOT EXISTS devices_fingerprint_uniq ON devices (manufacturer, model, build);
ALTER TABLE devices ADD CONSTRAINT devices_fingerprint_uniq UNIQUE USING INDEX devices_fingerprint_uniq;

CREATE TABLE IF NOT EXISTS devices_temp (
    manufacturer varchar,
    model varchar,
    os varchar,
    build varchar,
    cpu_description varchar,
    display_x integer,
    display_y integer,
    base_android_version varchar,
    dpi integer,
    build_os_device varchar
);

TRUNCATE devices_temp;

\COPY devices_temp FROM devices.csv DELIMITER ',' CSV HEADER;

INSERT INTO devices (manufacturer, model, build, cpu_description, display_x, display_y, base_android_version, dpi, build_os_device)
SELECT * FROM (
    SELECT manufacturer, model, build, cpu_description, display_x, display_y, cast(substring(base_android_version from '^\d\.\d') as double precision) as base_android_version, dpi, build_os_device
    FROM devices_temp
) inner_query
WHERE base_android_version > 6.0
ON CONFLICT ON CONSTRAINT devices_fingerprint_uniq
DO UPDATE SET
   cpu_description      = EXCLUDED.cpu_description,
   display_x            = EXCLUDED.display_x,
   display_y            = EXCLUDED.display_y,
   base_android_version = EXCLUDED.base_android_version,
   dpi                  = EXCLUDED.dpi,
   build_os_device      = EXCLUDED.build_os_device;

DROP TABLE IF EXISTS devices_temp;
