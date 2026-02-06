-- Auto-populate network_name for new stacks
-- Backfill existing stacks that have no network_name
UPDATE stacks SET network_name = 'devarch-' || name || '-net'
WHERE network_name IS NULL;
