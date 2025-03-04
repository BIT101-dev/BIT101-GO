-- 变量
create table variables
(
    id         bigserial
        primary key,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    obj        text not null,
    data       text
);

alter table variables
    owner to bit101;

create index idx_variables_deleted_at
    on variables (deleted_at);

create index idx_variables_obj
    on variables (obj);

create index idx_variables_created_at
    on variables (created_at);

create index idx_variables_updated_at
    on variables (updated_at);

INSERT INTO public.variables (id, created_at, updated_at, deleted_at, obj, data) VALUES (3, '2023-10-30 17:13:21.468441 +00:00', '2023-10-30 17:13:21.468441 +00:00', null, 'gallery_carousel', '[{"img":"https://cos.bit101.flwfdd.xyz/img/12cd23feea03706421d535ab6844b2d1.jpeg!low","url":"/gallery/"}]');
INSERT INTO public.variables (id, created_at, updated_at, deleted_at, obj, data) VALUES (2, '2023-05-16 04:09:56.159021 +00:00', '2025-02-12 18:26:55.848342 +00:00', null, 'billboard', '[{"title":"欢迎来到BIT101","text":"开业大酬宾！注册分享即送美好祝福一句！","url":"","img":""},{"title":"友情链接","text":"网络开拓者协会","url":"https://www.bitnp.net/","img":""},{"title":"登录","text":"注册登录账号 解锁更多功能","url":"/login/","img":""}]');
INSERT INTO public.variables (id, created_at, updated_at, deleted_at, obj, data) VALUES (1, '2023-05-16 04:09:56.157588 +00:00', '2025-02-12 18:27:30.072463 +00:00', null, 'carousel', '[{"img":"https://cos.bit101.flwfdd.xyz/img/12cd23feea03706421d535ab6844b2d1.jpeg!low","url":"/gallery/"},{"img":"https://cos.bit101.flwfdd.xyz/img/360a4e7b1e39ad6d2c09766c3d3f9e00.jpeg!low","url":""},{"img":"https://cos.bit101.flwfdd.xyz/img/09a50bf15ef77fac7a60ba0abced97e2.jpeg!low","url":""}]');


-- 用户身份
create table identities
(
    id         bigserial
        primary key,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    text       text not null,
    color      text
);

alter table identities
    owner to bit101;

create index idx_identities_deleted_at
    on identities (deleted_at);

create index idx_identities_created_at
    on identities (created_at);

create index idx_identities_updated_at
    on identities (updated_at);

INSERT INTO public.identities (id, created_at, updated_at, deleted_at, text, color) VALUES (0, '2023-10-30 17:08:07.611437 +00:00', '2023-10-30 17:08:07.611437 +00:00', null, '普通用户', null);
INSERT INTO public.identities (id, created_at, updated_at, deleted_at, text, color) VALUES (1, '2023-10-30 17:08:07.611437 +00:00', '2023-10-30 17:08:07.611437 +00:00', null, '管理员', '#FF6957');
INSERT INTO public.identities (id, created_at, updated_at, deleted_at, text, color) VALUES (2, '2023-10-30 17:08:07.611437 +00:00', '2023-10-30 17:08:07.611437 +00:00', null, '超级管理员', '#D63422');
INSERT INTO public.identities (id, created_at, updated_at, deleted_at, text, color) VALUES (3, '2023-10-30 17:08:07.611437 +00:00', '2023-10-30 17:08:07.611437 +00:00', null, '学生组织', '#00BCEB');
INSERT INTO public.identities (id, created_at, updated_at, deleted_at, text, color) VALUES (4, '2023-10-30 17:08:07.611437 +00:00', '2023-10-30 17:08:07.611437 +00:00', null, '社团', '#00D67D');
INSERT INTO public.identities (id, created_at, updated_at, deleted_at, text, color) VALUES (5, '2023-10-30 17:08:07.611437 +00:00', '2023-10-30 17:08:07.611437 +00:00', null, '非正式组织', '#F0BE51');
INSERT INTO public.identities (id, created_at, updated_at, deleted_at, text, color) VALUES (6, '2023-10-30 17:08:07.611437 +00:00', '2023-10-30 17:08:07.611437 +00:00', null, '机器人', '#8350EB');
INSERT INTO public.identities (id, created_at, updated_at, deleted_at, text, color) VALUES (7, '2023-10-30 17:08:07.611437 +00:00', '2023-10-30 17:08:07.611437 +00:00', null, '大会员', '#FF9A57');


-- 话廊声明
create table claims
(
    id         bigserial
        primary key,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    text       text not null
        unique
);

alter table claims
    owner to bit101;

create index idx_claims_deleted_at
    on claims (deleted_at);

create index idx_claims_created_at
    on claims (created_at);

create index idx_claims_updated_at
    on claims (updated_at);

INSERT INTO public.claims (id, created_at, updated_at, deleted_at, text) VALUES (0, '2023-10-30 17:03:08.956576 +00:00', '2023-10-30 17:03:08.956576 +00:00', null, '无声明');
INSERT INTO public.claims (id, created_at, updated_at, deleted_at, text) VALUES (1, '2023-10-30 17:03:08.956576 +00:00', '2023-10-30 17:03:08.956576 +00:00', null, '政治相关');
INSERT INTO public.claims (id, created_at, updated_at, deleted_at, text) VALUES (2, '2023-10-30 17:03:08.956576 +00:00', '2023-10-30 17:03:08.956576 +00:00', null, '性相关');
INSERT INTO public.claims (id, created_at, updated_at, deleted_at, text) VALUES (3, '2023-10-30 17:03:08.956576 +00:00', '2023-10-30 17:03:08.956576 +00:00', null, '商业广告');
INSERT INTO public.claims (id, created_at, updated_at, deleted_at, text) VALUES (4, '2023-10-30 17:03:08.956576 +00:00', '2023-10-30 17:03:08.956576 +00:00', null, '未经证实');


-- 举报类型
create table report_types
(
    id         bigserial
        primary key,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    text       text not null
);

alter table report_types
    owner to bit101;

create index idx_report_types_deleted_at
    on report_types (deleted_at);

create index idx_report_types_created_at
    on report_types (created_at);

create index idx_report_types_updated_at
    on report_types (updated_at);

INSERT INTO public.report_types (id, created_at, updated_at, deleted_at, text) VALUES (1, '2023-10-30 17:09:52.109756 +00:00', '2023-10-30 17:09:52.109756 +00:00', null, '政治敏感');
INSERT INTO public.report_types (id, created_at, updated_at, deleted_at, text) VALUES (2, '2023-10-30 17:09:52.109756 +00:00', '2023-10-30 17:09:52.109756 +00:00', null, '色情低俗');
INSERT INTO public.report_types (id, created_at, updated_at, deleted_at, text) VALUES (3, '2023-10-30 17:09:52.109756 +00:00', '2023-10-30 17:09:52.109756 +00:00', null, '人身攻击');
INSERT INTO public.report_types (id, created_at, updated_at, deleted_at, text) VALUES (4, '2023-10-30 17:09:52.109756 +00:00', '2023-10-30 17:09:52.109756 +00:00', null, '侵犯隐私');
INSERT INTO public.report_types (id, created_at, updated_at, deleted_at, text) VALUES (5, '2023-10-30 17:09:52.109756 +00:00', '2023-10-30 17:09:52.109756 +00:00', null, '散布谣言');
INSERT INTO public.report_types (id, created_at, updated_at, deleted_at, text) VALUES (6, '2023-10-30 17:09:52.109756 +00:00', '2023-10-30 17:09:52.109756 +00:00', null, '滥用产品');
INSERT INTO public.report_types (id, created_at, updated_at, deleted_at, text) VALUES (7, '2023-10-30 17:09:52.109756 +00:00', '2023-10-30 17:09:52.109756 +00:00', null, '其他');
