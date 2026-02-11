UPDATE services SET compose_overrides = jsonb_set(compose_overrides, '{build,context}', '"/workspace/config/flowstate"')
WHERE name IN ('flowstate-app', 'flowstate-queue', 'flowstate-reverb')
  AND compose_overrides -> 'build' ->> 'context' = '/workspace/apps/flowstate/deploy';
