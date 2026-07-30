-- Add missing columns to password_reset_history table
ALTER TABLE password_reset_history
ADD COLUMN IF NOT EXISTS initiated_by_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'pending',
ADD COLUMN IF NOT EXISTS form_attached BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS form_url TEXT,
ADD COLUMN IF NOT EXISTS notes TEXT,
ADD COLUMN IF NOT EXISTS token_expires_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS completed_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS organization_id BIGINT REFERENCES organizations(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS temporary_password_generated BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS email_sent BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS pdf_generated BOOLEAN DEFAULT FALSE;

-- Drop the old reset_by_user_id column if it exists (renamed to initiated_by_id)
ALTER TABLE password_reset_history DROP COLUMN IF EXISTS reset_by_user_id;

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_password_reset_history_initiated_by ON password_reset_history(initiated_by_id) WHERE initiated_by_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_password_reset_history_org ON password_reset_history(organization_id) WHERE organization_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_password_reset_history_status ON password_reset_history(status);

-- Update existing records to have default values
UPDATE password_reset_history
SET
    status = 'completed',
    form_attached = FALSE,
    temporary_password_generated = FALSE,
    email_sent = (reset_method = 'email'),
    pdf_generated = (reset_method = 'form')
WHERE status IS NULL;
