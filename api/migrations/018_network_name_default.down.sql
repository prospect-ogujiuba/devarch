-- Revert auto-populated network names
UPDATE stacks SET network_name = NULL WHERE network_name LIKE 'devarch-%-net';
