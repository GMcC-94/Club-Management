CREATE TABLE IF NOT EXISTS attendance_records (
    id SERIAL PRIMARY KEY,
    attendance_id INTEGER NOT NULL REFERENCES attendance(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    present BOOLEAN DEFAULT FALSE,
    class_id INTEGER REFERENCES classes(id),
    date DATE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(attendance_id, student_id)
);