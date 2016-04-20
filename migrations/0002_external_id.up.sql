CREATE EXTENSION "uuid-ossp";

alter table accounts add column external_id UUID not null default uuid_generate_v1mc();
alter table accounts add column role_arn character varying not null default '';

create table role_stacks (
  external_id UUID not null,
  customer_id UUID not null,
  stack_id character varying not null default '',
  stack_name character varying not null default '',
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone DEFAULT now() NOT NULL,
  primary key (customer_id, external_id)
);
