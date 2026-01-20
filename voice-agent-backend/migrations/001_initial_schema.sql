-- Voice Agent Database Schema
-- Run this in your Supabase SQL Editor

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table (identified by phone number)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255),
    email VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for phone lookups
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone_number);

-- Appointments table
CREATE TABLE IF NOT EXISTS appointments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_phone VARCHAR(20) NOT NULL,
    user_name VARCHAR(255),
    date_time TIMESTAMP WITH TIME ZONE NOT NULL,
    duration INTEGER DEFAULT 30, -- in minutes
    purpose TEXT,
    status VARCHAR(20) DEFAULT 'booked' CHECK (status IN ('booked', 'cancelled', 'completed')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for appointment queries
CREATE INDEX IF NOT EXISTS idx_appointments_user_phone ON appointments(user_phone);
CREATE INDEX IF NOT EXISTS idx_appointments_date_time ON appointments(date_time);
CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status);

-- Call summaries table
CREATE TABLE IF NOT EXISTS call_summaries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id VARCHAR(255) NOT NULL,
    user_phone VARCHAR(20),
    summary TEXT,
    appointments_booked JSONB DEFAULT '[]',
    user_preferences JSONB DEFAULT '[]',
    key_topics JSONB DEFAULT '[]',
    duration INTEGER, -- in seconds
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for summary lookups
CREATE INDEX IF NOT EXISTS idx_call_summaries_user_phone ON call_summaries(user_phone);
CREATE INDEX IF NOT EXISTS idx_call_summaries_session ON call_summaries(session_id);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_appointments_updated_at ON appointments;
CREATE TRIGGER update_appointments_updated_at
    BEFORE UPDATE ON appointments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Row Level Security (RLS) policies
-- Enable RLS on all tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE appointments ENABLE ROW LEVEL SECURITY;
ALTER TABLE call_summaries ENABLE ROW LEVEL SECURITY;

-- Create policies to allow all operations for the service role
-- Note: In production, you may want more restrictive policies

CREATE POLICY "Allow all operations for service role" ON users
    FOR ALL
    USING (true)
    WITH CHECK (true);

CREATE POLICY "Allow all operations for service role" ON appointments
    FOR ALL
    USING (true)
    WITH CHECK (true);

CREATE POLICY "Allow all operations for service role" ON call_summaries
    FOR ALL
    USING (true)
    WITH CHECK (true);

-- Grant permissions
GRANT ALL ON users TO anon, authenticated;
GRANT ALL ON appointments TO anon, authenticated;
GRANT ALL ON call_summaries TO anon, authenticated;

-- Sample data for testing (optional)
-- INSERT INTO users (phone_number, name) VALUES
--     ('+1234567890', 'John Doe'),
--     ('+0987654321', 'Jane Smith');
