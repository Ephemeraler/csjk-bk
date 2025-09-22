CREATE TABLE Model (
    ID INTEGER UNIQUE DEFAULT(1),
    version VARCHAR(20) NOT NULL,
    CONSTRAINT single_row CHECK (ID = 1)
);

INSERT INTO Model (version) VALUES ('1.0.0');

CREATE TABLE Alert (
	ID SERIAL NOT NULL PRIMARY KEY,
    fingerprint TEXT NOT NULL,
	status VARCHAR(50) NOT NULL,
    startsAt TIMESTAMPTZ NOT NULL,
    endsAt TIMESTAMPTZ NULL,
	generatorURL TEXT NOT NULL,
    responder TEXT NULL DEFAULT(''),
    operation TEXT NULL DEFAULT(''),
    CONSTRAINT alert_fingerprint_startsat UNIQUE (fingerprint,startsAt)
);

CREATE TABLE AlertLabel (
	ID SERIAL NOT NULL PRIMARY KEY,
    AlertID INT NOT NULL REFERENCES Alert (ID),
    Label VARCHAR(100) NOT NULL,
    Value VARCHAR(1000) NOT NULL
);

CREATE TABLE AlertAnnotation (
	ID SERIAL NOT NULL PRIMARY KEY,
    AlertID INT NOT NULL REFERENCES Alert (ID),
    Annotation VARCHAR(100) NOT NULL,
    Value VARCHAR(1000) NOT NULL,
    CONSTRAINT alertannotation_alertid_annotation UNIQUE (alertid,annotation)
);

CREATE INDEX idx_alert_startsat ON alert (startsat);
CREATE INDEX idx_alert_endsat ON alert (endsat);
CREATE INDEX idx_alertlabel_label_value_alertid ON alertlabel (label, value, alertid);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alertlabel_label ON alertlabel (label);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alertlabel_label_value ON public.alertlabel (label, value);

CREATE TABLE Applications (
    ID SERIAL NOT NULL PRIMARY KEY,
    Class VARCHAR(100) NOT NULL,
    -- ClusterID INT, -- 集群 ID
    State INT NOT NULL, -- 当前状态
    ApplyAt TIMESTAMPTZ NOT NULL DEFAULT now(), -- 申请日期
    Applier VARCHAR(100) NOT NULL, -- 申请用户 ID
    ReviewAt TIMESTAMPTZ NULL, -- 审核日期
    Reviewer VARCHAR(100) NULL, -- 审核用户 ID
    Decision text NULL, -- 审核结果
    Content JSON NOT NULL -- 申请内容
)