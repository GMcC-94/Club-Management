-- Users table (authentication)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'coach', 'treasurer')),
    must_change_password BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Students table
CREATE TABLE students (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    age INT,
    weight_range VARCHAR(50), -- e.g., "45-55kg"
    belt_level VARCHAR(50) NOT NULL, -- Red, White, Yellow, Orange, Green, Blue, Purple, Brown, Brown w/ stripe, Black
    fight_experience VARCHAR(20) CHECK (fight_experience IN ('novice', 'intermediate', 'advanced')),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Classes table (recurring schedule)
CREATE TABLE classes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL, -- e.g., "Monday Juniors"
    day_of_week INT NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6), -- 0=Sunday, 1=Monday, etc.
    start_time TIME NOT NULL,
    end_time TIME,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Student-Class mapping (which students are scheduled for which classes)
CREATE TABLE student_classes (
    id SERIAL PRIMARY KEY,
    student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    class_id INT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(student_id, class_id)
);

-- Attendance records
CREATE TABLE attendance (
    id SERIAL PRIMARY KEY,
    student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    class_id INT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('present', 'absent')),
    is_class_cancelled BOOLEAN DEFAULT FALSE,
    off_schedule BOOLEAN DEFAULT FALSE, -- TRUE if student attended a class they're not normally scheduled for
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(student_id, class_id, date)
);

-- Financial categories (custom, user-defined)
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expenditure')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(name, type)
);

-- Transactions (income and expenditure)
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expenditure')),
    amount DECIMAL(10, 2) NOT NULL,
    date DATE NOT NULL,
    description TEXT NOT NULL,
    category_id INT REFERENCES categories(id) ON DELETE SET NULL,
    payment_method VARCHAR(50) CHECK (payment_method IN ('cash', 'card', 'bank_transfer')),
    approved_by TEXT, -- Free text field
    receipt_filename VARCHAR(255), -- Stored file path
    created_by INT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- App settings
CREATE TABLE settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default settings
INSERT INTO settings (key, value) VALUES 
    ('attendance_warning_threshold', '75'),
    ('attendance_critical_threshold', '60');

-- Create indexes for common queries
CREATE INDEX idx_students_status ON students(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_students_belt ON students(belt_level) WHERE deleted_at IS NULL;
CREATE INDEX idx_attendance_date ON attendance(date) WHERE deleted_at IS NULL;
CREATE INDEX idx_attendance_student ON attendance(student_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_date ON transactions(date) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_type ON transactions(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_category ON transactions(category_id) WHERE deleted_at IS NULL;
