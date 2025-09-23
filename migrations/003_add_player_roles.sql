ALTER TABLE players
  ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';

CREATE INDEX IF NOT EXISTS idx_players_role ON players(role);

ALTER TABLE players DROP CONSTRAINT IF EXISTS check_player_role;
ALTER TABLE players ADD CONSTRAINT check_player_role
  CHECK (role IN ('user', 'admin'));

UPDATE players SET role = 'admin' WHERE id = 1 AND role = 'user';
