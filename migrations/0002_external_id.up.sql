alter table accounts add column external_id character varying not null default '';


create table role_stacks (
  external_id string not null default '',
  customer_id UUID not null,
  stack_id string not null default '',
  stack_name string not null default '',
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone DEFAULT now() NOT NULL,
  primary key (customer_id, external_id)
);
