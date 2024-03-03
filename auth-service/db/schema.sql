CREATE TABLE users (
    id         bigserial PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password   text NOT NULL,
    isverified BOOLEAN NOT NULL DEFAULT false,
    role       text CHECK (role IN ('Issuer', 'User', 'Verifier')) NOT NULL,
    otp        text NOT NULL
    CONSTRAINT valid_email CHECK (email ~ '^[a-zA-Z0-9.!#$%&''*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$')
);
