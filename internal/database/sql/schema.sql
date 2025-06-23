-- users
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- exercises
CREATE TABLE IF NOT EXISTS exercises (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    muscle_group VARCHAR(20) NOT NULL CHECK(muscle_group IN (
    'chest',
    'back',
    'legs',
    'core',
    'arms',
    'shoulders',
    'glutes'
    ))
);


-- workout_plans
CREATE TABLE IF NOT EXISTS workout_plans (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) NOT NULL,
    status VARCHAR(20) CHECK (status IN ('pending', 'completed', 'missed')),
    scheduled_date TIMESTAMP WITH TIME ZONE NOT NULL,
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- exercise_plans
CREATE TABLE IF NOT EXISTS exercise_plans (
    id SERIAL PRIMARY KEY,
    exercise_id INTEGER REFERENCES exercises(id) NOT NULL,
    workout_plan_id INTEGER REFERENCES workout_plans(id) NOT NULL,
    sets INT NOT NULL,
    repetitions INT NOT NULL,
    weights FLOAT NOT NULL,
    weight_unit VARCHAR(20) NOT NULL CHECK(weight_unit IN (
        'kg',
        'lbs',
        'other'
    ))
);
