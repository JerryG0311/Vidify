-- +goose Up
ALTER TABLE leads ADD COLUMN cta_id TEXT;
ALTER TABLE leads ADD COLUMN cta_type TEXT;
ALTER TABLE leads ADD COLUMN cta_time_seconds INTEGER;
ALTER TABLE leads ADD COLUMN cta_hero_text TEXT;

CREATE INDEX IF NOT EXISTS idx_leads_video_id ON leads(video_id);
CREATE INDEX IF NOT EXISTS idx_leads_cta_id ON leads(cta_id);




-- +goose Down
DROP INDEX IF EXISTS idx_leads_cta_id;
DROP INDEX IF EXISTS idx_leads_video_id;